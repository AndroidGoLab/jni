package client

import (
	"context"

	handlepb "github.com/xaionaro-go/jni/proto/handlestore"
	"google.golang.org/grpc"
)

// Client provides access to Android API services over gRPC.
// Currently only provides handle management. Android API service
// clients will be added once javagen produces matching Go wrappers.
type Client struct {
	handles handlepb.HandleStoreServiceClient
}

// NewClient creates a client from a gRPC connection.
func NewClient(cc grpc.ClientConnInterface) *Client {
	return &Client{
		handles: handlepb.NewHandleStoreServiceClient(cc),
	}
}

// ReleaseHandle releases a server-side JNI object handle.
func (c *Client) ReleaseHandle(ctx context.Context, handle int64) error {
	_, err := c.handles.ReleaseHandle(ctx, &handlepb.ReleaseHandleRequest{Handle: handle})
	return err
}
