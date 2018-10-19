package byproxy

type UseCommand struct {
	Environment string
	Proxies     []string
}

type Empty struct{}

type Receiver struct {
	server *Server
}

func (receiver *Receiver) Reload(config *Config, reply *Status) error {

	if status, e := receiver.server.Reload(config); e != nil {
		return e
	} else {
		*reply = *status
	}

	return nil
}

func (receiver *Receiver) Status(_ *Empty, reply *Status) error {
	*reply = *receiver.server.Status()
	return nil
}

func (receiver *Receiver) Use(command *UseCommand, reply *Status) error {

	if status, e := receiver.server.Patch(command.Environment, command.Proxies...); e != nil {
		return e
	} else {
		*reply = *status
	}

	return nil
}

func (receiver *Receiver) Stop(_, _ *Empty) error {
	return receiver.server.Close()
}
