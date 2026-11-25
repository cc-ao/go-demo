package greeter

import (
	"context"
	"strings"

	pb "github.com/cc-ao/go-demo/gen/helloworld/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the example greeting service.
type Server struct{ pb.UnimplementedGreeterServer }

func (Server) SayHello(_ context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name must not be empty")
	}
	return &pb.HelloReply{Message: "Hello, " + name + "!"}, nil
}
