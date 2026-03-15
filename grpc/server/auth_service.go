package server

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/xaionaro-go/jni/grpc/server/acl"
	"github.com/xaionaro-go/jni/grpc/server/certauth"
	pb "github.com/xaionaro-go/jni/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// PermissionRequestNotifier is called when a new permission request is
// created. The implementation should notify the device user (e.g. launch
// a dialog Activity or push a notification).
type PermissionRequestNotifier func(requestID int64, clientID string, methods []string)

// AuthServiceServer implements pb.AuthServiceServer.
type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
	CA                  *certauth.CA
	Store               *acl.Store
	OnPermissionRequest PermissionRequestNotifier
}

// Register handles unauthenticated registration: it signs the submitted CSR
// and registers the resulting client in the ACL store.
func (s *AuthServiceServer) Register(
	_ context.Context,
	req *pb.RegisterRequest,
) (*pb.RegisterResponse, error) {
	certPEM, err := s.CA.SignCSR([]byte(req.GetCsrPem()))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "sign CSR: %v", err)
	}

	// Parse the signed cert to extract CN and fingerprint.
	cert, err := certauth.ParseCertPEM(certPEM)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "parse signed cert: %v", err)
	}

	fingerprint := fmt.Sprintf("sha256:%x", sha256.Sum256(cert.Raw))

	if err := s.Store.RegisterClient(cert.Subject.CommonName, string(certPEM), fingerprint); err != nil {
		return nil, status.Errorf(codes.Internal, "register client: %v", err)
	}

	return &pb.RegisterResponse{
		ClientCertPem: string(certPEM),
		CaCertPem:     string(s.CA.CertPEM()),
	}, nil
}

// RequestPermission creates a pending permission request for the calling
// client (identified via mTLS peer certificate CN).
func (s *AuthServiceServer) RequestPermission(
	ctx context.Context,
	req *pb.RequestPermissionRequest,
) (*pb.RequestPermissionResponse, error) {
	clientID, err := clientIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	reqID, err := s.Store.CreateRequest(clientID, req.GetMethods())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create request: %v", err)
	}

	// Notify the device user (launch approval dialog).
	if s.OnPermissionRequest != nil {
		s.OnPermissionRequest(reqID, clientID, req.GetMethods())
	}

	return &pb.RequestPermissionResponse{
		RequestId: fmt.Sprintf("%d", reqID),
		Status:    "pending",
	}, nil
}

// ListMyPermissions returns all granted method patterns for the calling
// client (identified via mTLS peer certificate CN).
func (s *AuthServiceServer) ListMyPermissions(
	ctx context.Context,
	_ *pb.ListMyPermissionsRequest,
) (*pb.ListMyPermissionsResponse, error) {
	clientID, err := clientIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	grants, err := s.Store.ListGrants(clientID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list grants: %v", err)
	}

	methods := make([]string, len(grants))
	for i, g := range grants {
		methods[i] = g.MethodPattern
	}

	return &pb.ListMyPermissionsResponse{GrantedMethods: methods}, nil
}

// clientIDFromContext extracts the client CN from the TLS peer certificate.
func clientIDFromContext(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no peer info")
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok || len(tlsInfo.State.PeerCertificates) == 0 {
		return "", status.Error(codes.Unauthenticated, "no client certificate")
	}

	return tlsInfo.State.PeerCertificates[0].Subject.CommonName, nil
}
