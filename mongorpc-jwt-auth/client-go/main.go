package main

import (
	"context"
	"time"

	"github.com/mongorpc/mongorpc/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// gRPC server address
	address = "localhost:27051"
)

func main() {
	interceptor := &ClientInterceptor{
		accessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJ1aWQiOiI1YjEwODcxZS0zYmMwLTExZWMtOGQzZC0wMjQyYWMxMzAwMDMifQ.PlVY78uCmkzzKVG3tOvQu9aeXnMOImUzG6b_lygHH2U",
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(address,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(interceptor.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(interceptor.StreamClientInterceptor),
	)
	if err != nil {
		logrus.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create a new client
	c := proto.NewMongoRPCClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	r, err := c.ListCollections(ctx, &proto.ListCollectionsRequest{
		Database: "sample_mflix",
	})
	if err != nil {
		logrus.Fatalf("could not get collection: %v", err)
	}
	logrus.Printf("Collection: %s", r.Collections)
}

type ClientInterceptor struct {
	accessToken string
}

// UnaryClientInterceptor is a gRPC interceptor that adds the access token to the request
func (interceptor *ClientInterceptor) UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	if interceptor.accessToken != "" {
		md.Set("authorization", interceptor.accessToken)
	}

	ctx = metadata.NewOutgoingContext(ctx, md)

	return invoker(ctx, method, req, reply, cc, opts...)
}

// StreamClientInterceptor is a gRPC interceptor that adds the access token to the request
func (interceptor ClientInterceptor) StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	if interceptor.accessToken != "" {
		md.Set("authorization", interceptor.accessToken)
	}

	ctx = metadata.NewOutgoingContext(ctx, md)

	return streamer(ctx, desc, cc, method, opts...)
}
