package byproxy

import (
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

func (remote *Remote) Reload(config *Config) (*Status, error) {
	status := &Status{}
	return status, remote.client.Call("Receiver.Reload", config, status)
}

func (remote *Remote) Status() (*Status, error) {
	status := &Status{}
	return status, remote.client.Call("Receiver.Status", &Empty{}, status)
}

func (remote *Remote) Use(env string, proxies ...string) (*Status, error) {
	status := &Status{}
	command := &UseCommand{Environment: env, Proxies: proxies}
	return status, remote.client.Call("Receiver.Use", command, status)
}

func (remote *Remote) Stop() error {
	return remote.client.Call("Receiver.Stop", &Empty{}, &Empty{})
}
