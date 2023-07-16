package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"

	"github.com/27182818284590/proto/foo"

	"github.com/27182818284590/udsrpc"
)

const (
	protocol = "unix"
	sockAddr = "/tmp/foo.sock"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := os.RemoveAll(sockAddr); err != nil {
		log.Fatalf("Remove socket: %s", err)
	}

	lis, err := net.Listen(protocol, sockAddr)
	if err != nil {
		log.Fatalf("Listen: %s", err)
	}

	if err := os.Chmod(sockAddr, 0o770); err != nil {
		log.Fatalf("chmod: %s", err)
	}

	udsRPC := udsrpc.New()

	s := grpc.NewServer()
	foo.RegisterFooServer(s, udsRPC)

	errChan := make(chan error)

	go func() {
		log.Printf("Server started %q", sockAddr)
		errChan <- s.Serve(lis)
	}()

	select {
	case err := <-errChan:
		log.Fatalf("Server failed: %s", err)
	case <-ctx.Done():
		log.Print("Shutdown")
	}

	s.GracefulStop()
}
