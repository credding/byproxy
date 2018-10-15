package server

import (
	"fmt"
	"github.com/credding/byproxy/pkg/util"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
)

type ByProxy struct {
	config   *util.Config
	remote   *rpc.Server
	routes   map[string]*Route
	handlers map[string]map[string]http.Handler
	server   *http.Server
}

func NewByProxy(config *util.Config) (*ByProxy, error) {

	proxy := &ByProxy{
		remote: rpc.NewServer(),
		server: &http.Server{},
	}

	if e := proxy.remote.Register(&Remote{proxy}); e != nil {
		return nil, e
	}
	proxy.remote.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	if e := proxy.Reload(config); e != nil {
		return nil, e
	}

	return proxy, nil
}

func (proxy *ByProxy) ListenAndServe() error {

	global := proxy.config.Global
	addr := fmt.Sprintf("%s:%d", global.Addr, global.Port)

	listener, e := net.Listen("tcp", addr)
	if e != nil {
		return e
	}

	if e = proxy.server.Serve(listener); e != nil {
		return e
	}

	return nil
}

func (proxy *ByProxy) Reload(config *util.Config) error {

	proxy.config = config

	mux := http.NewServeMux()
	mux.Handle("/", proxy.remote)

	proxy.routes = make(map[string]*Route)

	for _, proxyConfig := range config.Proxies {
		handler := NewRoute()
		proxy.routes[proxyConfig.Name] = handler
		mux.Handle(proxyConfig.Hostname+"/", handler)
	}

	servers := make(map[string]*util.Server)

	for _, serverConfig := range config.Servers {
		servers[serverConfig.Name] = serverConfig
	}

	proxy.handlers = make(map[string]map[string]http.Handler)

	for _, environmentConfig := range config.Environments {

		handlers := make(map[string]http.Handler)
		proxy.handlers[environmentConfig.Name] = handlers

		for route, server := range environmentConfig.Mappings {
			serverConfig := servers[server]
			baseUrl, e := url.Parse(serverConfig.BaseUrl)
			if e != nil {
				return e
			}
			handler := NewFilterProxy(baseUrl, nil)
			handlerHostname := route + "-" + environmentConfig.Name + "." + config.Global.Hostname + "/"
			handlers[route] = handler
			mux.Handle(handlerHostname, handler)
		}
	}

	proxy.server.Handler = mux

	return nil
}

func (proxy *ByProxy) LoadMappings(mappings map[string]string) {

	for route, environment := range mappings {

		handlers, ok := proxy.handlers[environment]
		if !ok {
			continue
		}

		if handler, ok := handlers[route]; ok {
			proxy.routes[route].Patch(environment, handler)
		}
	}
}

func (proxy *ByProxy) Patch(environment string, proxies ...string) {

	handlers, ok := proxy.handlers[environment]
	if !ok {
		return
	}

	if proxies != nil {
		for _, route := range proxies {
			if handler, ok := handlers[route]; ok {
				proxy.routes[route].Patch(environment, handler)
			}
		}
	} else {
		for route, handler := range handlers {
			proxy.routes[route].Patch(environment, handler)
		}
	}
}

func (proxy *ByProxy) Status() *Status {

	mappings := make(map[string]string, len(proxy.routes))
	environments := make([]string, len(proxy.handlers))

	for route, handler := range proxy.routes {
		mappings[route] = handler.Environment
	}
	i := 0
	for environment := range proxy.handlers {
		environments[i] = environment
		i = i + 1
	}

	return &Status{mappings, environments}
}

func (proxy *ByProxy) Close() error {
	return proxy.server.Close()
}
