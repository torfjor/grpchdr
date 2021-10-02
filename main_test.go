package grpchdr_test

import (
	"context"
	"grpchdr/pb"
	"testing"

	"golang.org/x/net/nettest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type emptyFn func(ctx context.Context, req *pb.EmptyRequest) (*pb.EmptyResponse, error)

type emptyService struct {
	emptyFn
	pb.UnimplementedEmptyServiceServer
}

func (e emptyService) Empty(ctx context.Context, req *pb.EmptyRequest) (*pb.EmptyResponse, error) {
	return e.emptyFn(ctx, req)
}

func newClientServerPair(t *testing.T, svc pb.EmptyServiceServer) (*grpc.Server, pb.EmptyServiceClient) {
	srv := grpc.NewServer()
	pb.RegisterEmptyServiceServer(srv, svc)

	lis, err := nettest.NewLocalListener("unix")
	if err != nil {
		t.Fatalf("nettest.NewLocalListener() = %v", err)
	}

	go func() {
		if err := srv.Serve(lis); err != grpc.ErrServerStopped {
			panic(err)
		}
	}()

	conn, err := grpc.Dial("unix:///"+lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("grpc.Dial() = %v", err)
	}
	c := pb.NewEmptyServiceClient(conn)

	return srv, c
}

func TestMetadata_In_Context(t *testing.T) {
	srv, client := newClientServerPair(t, emptyService{
		emptyFn: func(ctx context.Context, req *pb.EmptyRequest) (*pb.EmptyResponse, error) {
			if md, ok := metadata.FromIncomingContext(ctx); !ok {
				t.Errorf("no metadata in context")
			} else if len(md.Get("x-foo")) < 1 {
				t.Errorf("no values in \"foo\"")
			}
			return &pb.EmptyResponse{}, nil
		},
	})
	defer srv.Stop()

	ctx := metadata.AppendToOutgoingContext(context.Background(), "x-foo", "bars")
	if _, err := client.Empty(ctx, &pb.EmptyRequest{}); err != nil {
		t.Errorf("client.Empty() = %v", err)
	}
}
