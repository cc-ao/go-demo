package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	pb "github.com/cc-ao/go-demo/gen/helloworld/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func run() error {
	addr := flag.String("addr", "127.0.0.1:50051", "gRPC server address")
	name := flag.String("name", "World", "name to greet")
	timeout := flag.Duration("timeout", 3*time.Second, "RPC deadline (must be positive)")
	flag.Parse()
	if *timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	reply, err := pb.NewGreeterClient(conn).SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		return fmt.Errorf("SayHello: %w", err)
	}
	fmt.Println(reply.GetMessage())
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
