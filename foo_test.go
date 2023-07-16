package udsrpc

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"

	"github.com/27182818284590/proto/foo"
)

const (
	protocol = "unix"
	sockAddr = "/tmp/b.sock"
	tcpAddr  = "localhost:12345"
)

var (
	tcpClient foo.FooClient
	udsClient foo.FooClient
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := os.RemoveAll(sockAddr); err != nil {
		log.Fatalf("Remove socket: %s", err)
	}

	udsLis, err := net.Listen("unix", sockAddr)
	if err != nil {
		log.Fatalf("Listen: %s", err)
	}

	tcpLis, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("Listen: %s", err)
	}

	if err := os.Chmod(sockAddr, 0o770); err != nil {
		log.Fatalf("chmod: %s", err)
	}

	fooRPC := New()

	s1 := grpc.NewServer()
	s2 := grpc.NewServer()

	foo.RegisterFooServer(s1, fooRPC)
	foo.RegisterFooServer(s2, fooRPC)

	udsDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return net.Dial("unix", addr)
	}

	udsOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithContextDialer(udsDialer),
	}

	udsConn, err := grpc.Dial(sockAddr, udsOpts...)
	if err != nil {
		log.Fatalf("UDS foo connection: %s", err)
	}
	defer udsConn.Close()

	udsClient = foo.NewFooClient(udsConn)

	tcpDialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return net.Dial("tcp", addr)
	}

	tcpOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithContextDialer(tcpDialer),
	}

	tcpConn, err := grpc.Dial(tcpAddr, tcpOpts...)
	if err != nil {
		log.Fatalf("TCP foo connection: %s", err)
	}
	defer tcpConn.Close()

	tcpClient = foo.NewFooClient(tcpConn)

	errChan := make(chan error)

	go func() {
		log.Printf("UDS Server started %q", sockAddr)
		errChan <- s1.Serve(udsLis)
	}()

	go func() {
		log.Printf("TCP Server started %q", tcpAddr)
		errChan <- s2.Serve(tcpLis)
	}()

	go func() {
		select {
		case err := <-errChan:
			log.Fatalf("Server failed: %s", err)
		case <-ctx.Done():
			log.Print("Shutdown")
			s1.GracefulStop()
			s2.GracefulStop()
		}
	}()

	os.Exit(m.Run())
}

var (
	resp *foo.FooResp
	err  error
)

func BenchmarkFooUDS(b *testing.B) {
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		req := &foo.FooReq{
			User: int64(i),
		}
		resp, err = udsClient.DoFoo(ctx, req)
		if err != nil {
			b.Fatalf("DoFoo: %s", err)
		}
	}
}

func BenchmarkFooTCP(b *testing.B) {
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		req := &foo.FooReq{
			User: int64(i),
		}
		resp, err = tcpClient.DoFoo(ctx, req)
		if err != nil {
			b.Fatalf("DoFoo: %s", err)
		}
	}
}
