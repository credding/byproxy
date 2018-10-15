package cli

import (
	"github.com/credding/byproxy/pkg/server"
	"net/rpc"
)

type Remote struct {
	client *rpc.Client
}

func NewRemote(addr string) (*Remote, error) {
	if client, e := rpc.DialHTTP("tcp", addr); e == nil {
		return &Remote{client}, nil
	} else {
		return nil, e
	}
}

func (remote *Remote) Reload() (*server.Status, error) {
	status := &server.Status{}
	return status, remote.client.Call("Remote.Reload", "", status)
}

func (remote *Remote) Status() (*server.Status, error) {
	status := &server.Status{}
	return status, remote.client.Call("Remote.Status", &server.Empty{}, status)
}

func (remote *Remote) Use(env string, proxies ...string) (*server.Status, error) {
	status := &server.Status{}
	command := &server.UseCommand{Environment: env, Proxies: proxies}
	return status, remote.client.Call("Remote.Use", command, status)
}

func (remote *Remote) Stop() error {
	return remote.client.Call("Remote.Stop", &server.Empty{}, &server.Empty{})
}
