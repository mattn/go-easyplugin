package easyplugin

import (
	"bufio"
	"io"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Plugin struct {
	Name   string
	cmd    *exec.Cmd
	stderr io.ReadCloser
	stdout io.ReadCloser
	stdin  io.WriteCloser
}

type PluginSystem struct {
	Name    string
	plugins []*Plugin
	f       func(string)
	input   chan string
	tmp     chan string
}

func New(name string) (*PluginSystem, error) {
	home := os.Getenv("HOME")
	p := ""
	if home == "" && runtime.GOOS == "windows" {
		p = filepath.Join(os.Getenv("APPDATA"), name, "plugins")
	} else {
		p = filepath.Join(home, ".config", name, "plugins")
	}
	err := os.MkdirAll(p, 755)
	if err != nil {
		return nil, err
	}
	dir, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	ps := &PluginSystem{
		input: make(chan string),
	}
	runtime.SetFinalizer(ps, func(ps *PluginSystem) {
		ps.Unload()
	})

	for _, fi := range fis {
		if runtime.GOOS == "windows" && !strings.HasSuffix(fi.Name(), ".exe") {
			continue
		} else if fi.Mode().Perm()&0700 == 0 {
			continue
		}
		cmd := exec.Command(filepath.Join(p, fi.Name()), "-d")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nil, err
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		err = cmd.Start()
		if err != nil {
			return nil, err
		}
		name := filepath.Base(fi.Name())
		if ext := filepath.Ext(name); ext != "" {
			name = name[:len(name)-len(ext)]
		}
		plugin := &Plugin{
			Name:   name,
			cmd:    cmd,
			stdin:  stdin,
			stdout: stdout,
			stderr: stderr,
		}
		ps.plugins = append(ps.plugins, plugin)
		if strings.HasPrefix(plugin.Name, "client-") {
			go func() {
				s := bufio.NewScanner(plugin.stdout)
				for s.Scan() {
					res := s.Text()
					ps.input <- res
				}
			}()
		}
	}
	return ps, nil
}

func (ps *PluginSystem) Stop() {
	close(ps.input)
}

func (ps *PluginSystem) ListenAndServe() {
	for data := range ps.input {
		ps.f(data)
	}
}

func (ps *PluginSystem) Handle(f func(data string)) {
	ps.f = f
}

func (ps *PluginSystem) Unload() {
	for _, p := range ps.plugins {
		p.cmd.Process.Kill()
	}
}

type ReadWriteCloser struct {
	io.Reader
	io.Writer
}

func (c *ReadWriteCloser) Close() error {
	return nil
}

func (ps *PluginSystem) Call(method string, args interface{}) error {
	for _, plugin := range ps.plugins {
		if !strings.HasPrefix(plugin.Name, "server-") {
			continue
		}
		rwc := &ReadWriteCloser{plugin.stdout, plugin.stdin}
		client := jsonrpc.NewClient(rwc)
		var res interface{}
		client.Call(method, args, &res)
		client.Close()
	}
	return nil
}

func (ps *PluginSystem) CallFor(name string, method string, res interface{}, args interface{}) error {
	for _, plugin := range ps.plugins {
		if plugin.Name != name {
			continue
		}
		rwc := &ReadWriteCloser{plugin.stdout, plugin.stdin}
		client := jsonrpc.NewClient(rwc)
		err := client.Call(method, args, &res)
		client.Close()
		return err
	}
	return nil
}
