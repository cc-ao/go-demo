package greeter

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/cc-ao/go-demo/gen/helloworld/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestSayHello(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, Server{})
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Errorf("Serve: %v", err)
		}
	}()
	t.Cleanup(func() { server.Stop(); listener.Close() })
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	client := pb.NewGreeterClient(conn)
	for _, tc := range []struct {
		name, input, want string
		code              codes.Code
	}{
		{"default", "World", "Hello, World!", codes.OK},
		{"unicode", "小明", "Hello, 小明!", codes.OK},
		{"trim", " Alice ", "Hello, Alice!", codes.OK},
		{"inner space", "Alice Smith", "Hello, Alice Smith!", codes.OK},
		{"empty", "", "", codes.InvalidArgument},
		{"whitespace", " \t\n\u3000", "", codes.InvalidArgument},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			reply, err := client.SayHello(ctx, &pb.HelloRequest{Name: tc.input})
			if status.Code(err) != tc.code {
				t.Fatalf("code = %v, want %v: %v", status.Code(err), tc.code, err)
			}
			if tc.code == codes.OK && reply.GetMessage() != tc.want {
				t.Fatalf("message = %q, want %q", reply.GetMessage(), tc.want)
			}
		})
	}
}
