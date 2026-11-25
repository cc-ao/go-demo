package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/cc-ao/go-demo/gen/helloworld/v1"
	"github.com/cc-ao/go-demo/internal/greeter"
	"google.golang.org/grpc"
)

func run() error {
	addr := flag.String("addr", "127.0.0.1:50051", "TCP listen address")
	flag.Parse()
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", *addr, err)
	}
	defer listener.Close()
	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, greeter.Server{})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	log.Printf("listening on %s", listener.Addr())
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		done := make(chan struct{})
		go func() { server.GracefulStop(); close(done) }()
		timer := time.NewTimer(3 * time.Second)
		defer timer.Stop()
		select {
		case <-done:
		case <-timer.C:
			server.Stop()
		}
		return <-result
	}
}

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
