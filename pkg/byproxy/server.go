package byproxy

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"path"
	"syscall"
)

type Server struct {
	proxy *ByProxy
	http  *http.Server
	rpc   *rpc.Server
}

func NewServer(config *Config) (*Server, error) {

	if proxy, e := NewByProxy(config); e != nil {
		return nil, e
	} else {

		serveMux := http.NewServeMux()
		rpcServer := rpc.NewServer()

		httpServer := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", config.Global.Addr, config.Global.Port),
			Handler: serveMux,
		}

		server := &Server{
			proxy: proxy,
			http:  httpServer,
			rpc:   rpcServer,
		}

		serveMux.Handle("/", proxy)
		serveMux.Handle(rpc.DefaultRPCPath, server.rpc)

		rpcServer.Register(&Receiver{server: server})

		server.reloadState()
		server.saveState()

		return server, nil
	}
}

func (server *Server) ListenAndServe() error {
	return server.http.ListenAndServe()
}

func (server *Server) Reload(config *Config) (*Status, error) {
	if e := server.proxy.Reload(config); e != nil {
		return nil, e
	}
	server.reloadState()
	return server.saveState(), nil
}

func (server *Server) Patch(environment string, proxies ...string) (*Status, error) {
	if e := server.proxy.Patch(environment, proxies...); e != nil {
		return nil, e
	}
	return server.saveState(), nil
}

func (server *Server) Status() *Status {
	return server.proxy.Status()
}

func (server *Server) Close() error {
	return server.http.Close()
}

func (server *Server) reloadState() {

	location := path.Join(userHomeDir(), ".byp/state")

	bytes, e := ioutil.ReadFile(location)
	if e != nil {
		if e, ok := e.(*os.PathError); !ok || e.Err == syscall.ENOENT {
			log.Printf("error loading state: %s\n", e)
		}
		return
	}

	state := make(map[string]string)
	if e := yaml.Unmarshal(bytes, &state); e != nil {
		log.Printf("error loading state: %s\n", e)
		return
	}

	server.proxy.PatchState(state)
}

func (server *Server) saveState() *Status {

	location := path.Join(userHomeDir(), ".byp/state")
	status := server.proxy.Status()

	bytes, _ := yaml.Marshal(status.Mappings)
	if e := ioutil.WriteFile(location, bytes, 0644); e != nil {
		log.Printf("error saving state: %s\n", e)
	}

	return status
}
