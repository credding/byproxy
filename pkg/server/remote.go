package server

import (
	"fmt"
	"github.com/credding/byproxy/pkg/util"
)

type Remote struct {
	proxy *ByProxy
}

func (remote *Remote) Reload(location string, reply *Status) error {

	if config, e := util.LoadConfig(location); e == nil {
		if e := remote.proxy.Reload(config); e != nil {
			return e
		}
	} else {
		return e
	}

	if mappings, e := util.LoadMappings(); mappings != nil {
		remote.proxy.LoadMappings(mappings)
	} else if e != nil {
		fmt.Println(e)
	}

	*reply = *remote.proxy.Status()
	return nil
}

func (remote *Remote) Status(_ *Empty, reply *Status) error {
	*reply = *remote.proxy.Status()
	return nil
}

func (remote *Remote) Use(command *UseCommand, reply *Status) error {

	remote.proxy.Patch(command.Environment, command.Proxies...)
	status := remote.proxy.Status()

	if e := util.WriteMappings(status.Mappings); e != nil {
		return e
	}

	*reply = *status
	return nil
}

func (remote *Remote) Stop(_, _ *Empty) error {
	return remote.proxy.Close()
}
