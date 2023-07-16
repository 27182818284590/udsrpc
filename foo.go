package udsrpc

import (
	"context"

	"github.com/27182818284590/proto/foo"
)

type UDSRPC struct {
	foo.UnimplementedFooServer
}

func New() *UDSRPC {
	return &UDSRPC{}
}

func (ur *UDSRPC) DoFoo(ctx context.Context, req *foo.FooReq) (*foo.FooResp, error) {
	println(req.Text)

	return &foo.FooResp{
		User: req.User,
	}, nil
}
