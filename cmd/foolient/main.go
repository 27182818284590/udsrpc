package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"

	"github.com/27182818284590/proto/foo"
)

const (
	protocol = "unix"
	sockAddr = "/tmp/foo.sock"
)

var transmission = []struct {
	UserID int64
	Text   string
}{
	{42, "Hello there 42"},
	{13, "Row 13, seat F"},
	{11, "Eleven"},
	{97, "Opa"},
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return net.Dial(protocol, addr)
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithContextDialer(dialer),
	}

	conn, err := grpc.Dial(sockAddr, opts...)
	if err != nil {
		log.Fatalf("foo connection: %s", err)
	}
	defer conn.Close()

	client := foo.NewFooClient(conn)

	for _, t := range transmission {
		req := &foo.FooReq{
			User: t.UserID,
			Text: t.Text,
		}

		resp, err := client.DoFoo(ctx, req)
		if err != nil {
			log.Fatalf("do foo: %s", err)
		}

		println(resp.User)
	}
}
