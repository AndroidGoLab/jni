package server

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNoAuth(t *testing.T) {
	auth := NoAuth{}
	if err := auth.Authorize(context.Background(), "/pkg.Service/Method"); err != nil {
		t.Fatalf("NoAuth.Authorize returned error: %v", err)
	}
}

func TestTokenAuth_Valid(t *testing.T) {
	auth := TokenAuth{Token: "secret123"}
	md := metadata.Pairs("authorization", "Bearer secret123")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	if err := auth.Authorize(ctx, "/pkg.Service/Method"); err != nil {
		t.Fatalf("TokenAuth.Authorize with valid token returned error: %v", err)
	}
}

func TestTokenAuth_MissingMetadata(t *testing.T) {
	auth := TokenAuth{Token: "secret123"}
	err := auth.Authorize(context.Background(), "/pkg.Service/Method")
	if err == nil {
		t.Fatal("expected error for missing metadata, got nil")
	}
	if s, ok := status.FromError(err); !ok || s.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", err)
	}
}

func TestTokenAuth_MissingAuthorization(t *testing.T) {
	auth := TokenAuth{Token: "secret123"}
	md := metadata.Pairs("other-key", "value")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := auth.Authorize(ctx, "/pkg.Service/Method")
	if err == nil {
		t.Fatal("expected error for missing authorization, got nil")
	}
	if s, ok := status.FromError(err); !ok || s.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", err)
	}
}

func TestTokenAuth_InvalidToken(t *testing.T) {
	auth := TokenAuth{Token: "secret123"}
	md := metadata.Pairs("authorization", "Bearer wrong-token")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := auth.Authorize(ctx, "/pkg.Service/Method")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
	if s, ok := status.FromError(err); !ok || s.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", err)
	}
}

func TestTokenAuth_WrongFormat(t *testing.T) {
	auth := TokenAuth{Token: "secret123"}
	md := metadata.Pairs("authorization", "secret123")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := auth.Authorize(ctx, "/pkg.Service/Method")
	if err == nil {
		t.Fatal("expected error for token without Bearer prefix, got nil")
	}
	if s, ok := status.FromError(err); !ok || s.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", err)
	}
}
