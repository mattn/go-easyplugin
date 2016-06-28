package main

import (
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
)

type Args struct {
	A, B int
}

type Reply struct {
	C int
}

type Calc int

func (t *Calc) Add(args *Args, reply *Reply) error {
	reply.C = args.A + args.B
	return nil
}

type ReadWriteCloser struct {
	io.Reader
	io.Writer
}

func (rwc *ReadWriteCloser) Close() error {
	return nil
}

func main() {
	calc := new(Calc)
	server := rpc.NewServer()
	server.Register(calc)
	rwc := &ReadWriteCloser{os.Stdin, os.Stdout}
	for {
		server.ServeCodec(jsonrpc.NewServerCodec(rwc))
	}
}
