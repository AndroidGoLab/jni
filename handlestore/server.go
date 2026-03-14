package handlestore

import (
	"context"

	"github.com/xaionaro-go/jni"
	pb "github.com/xaionaro-go/jni/proto/handlestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the HandleStoreService gRPC server.
type Server struct {
	pb.UnimplementedHandleStoreServiceServer
	VM      *jni.VM
	Handles *HandleStore
}

// ReleaseHandle releases a previously-stored JNI global reference.
func (s *Server) ReleaseHandle(_ context.Context, req *pb.ReleaseHandleRequest) (*pb.ReleaseHandleResponse, error) {
	handle := req.GetHandle()
	if handle == 0 {
		return &pb.ReleaseHandleResponse{}, nil
	}
	if s.Handles.Get(handle) == nil {
		return nil, status.Errorf(codes.NotFound, "handle %d not found", handle)
	}
	if err := s.VM.Do(func(env *jni.Env) error {
		s.Handles.Release(env, handle)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "release handle: %v", err)
	}
	return &pb.ReleaseHandleResponse{}, nil
}
