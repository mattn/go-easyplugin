# go-easyplugin

Easy Plugin System for Go

**VERY VERY EXPERIMENTAL**

## Usage

### Application

```go
ps, err := easyplugin.New("foobar")
defer ps.Unload()
```

The plugin applications located in your `~/.config/foobar/plugins/xxx` will be spawned.

### Client Plugin

If the plugin name is `client-xxx`, it works as notificator. So the implementation will be:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("foo")
	time.Sleep(1 * time.Second)
	fmt.Println("bar")
	time.Sleep(1 * time.Second)
	fmt.Println("baz")
	time.Sleep(1 * time.Second)
}
```

You can handle notifications like below

```go
ps.Handle(func(data string) {
	log.Println(data)
})
ps.ListenAndServe()
```

### Server Plugin

If the plugin name is `server-xxx`, it works as server. So the implementation will be:

```go
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
```

Server plugin can talk to main application with JSON-RPC. You can call the method in server plugin like below

```go
err = ps.CallFor("server-app1", "Calc.Add", &res, struct {
	A, B int
}{1, 3})
log.Println(res) // 4
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a mattn)
