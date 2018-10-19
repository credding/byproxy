package byproxy

import (
	"fmt"
	"net/http"
	"net/url"
)

type Status struct {
	Proxies      []string
	Environments []string
	Mappings     map[string]string
}

type ByProxy struct {
	proxyNames       []string
	environmentNames []string
	routes           routeMap
	environments     environmentMap
	serveMux         *http.ServeMux
}

type routeMap map[string]*Route
type serverMap map[string]*ServerConfig
type handlerMap map[string]http.Handler
type environmentMap map[string]handlerMap

func NewByProxy(config *Config) (*ByProxy, error) {

	proxy := &ByProxy{}

	if e := proxy.Reload(config); e != nil {
		return nil, e
	}

	return proxy, nil
}

func (proxy *ByProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	proxy.serveMux.ServeHTTP(rw, req)
}

func (proxy *ByProxy) Reload(config *Config) error {

	if proxyNames, routes, e := initRoutes(config.Proxies); e != nil {
		return e
	} else if servers, e := initServers(config.Servers); e != nil {
		return e
	} else if environmentNames, environments, e := initEnvironments(config.Environments, routes, servers); e != nil {
		return e
	} else {

		state := proxy.Mappings()

		proxy.proxyNames = proxyNames
		proxy.environmentNames = environmentNames

		proxy.routes = routes
		proxy.environments = environments
		proxy.serveMux = initServeMux(config, routes, environments)

		proxy.PatchState(state)
	}

	return nil
}

func (proxy *ByProxy) Patch(environment string, proxies ...string) error {

	handlers, ok := proxy.environments[environment]
	if !ok {
		return fmt.Errorf("no such environment: %s", environment)
	}

	if proxies != nil {
		for _, name := range proxies {
			if _, ok := proxy.routes[name]; !ok {
				return fmt.Errorf("no such proxy: %s", name)
			}
		}
		for _, name := range proxies {
			if handler, ok := handlers[name]; ok {
				proxy.routes[name].Patch(environment, handler)
			}
		}
	} else {
		for name, handler := range handlers {
			proxy.routes[name].Patch(environment, handler)
		}
	}

	return nil
}

func (proxy *ByProxy) PatchState(state map[string]string) {

	for route, environment := range state {

		handlers, ok := proxy.environments[environment]
		if !ok {
			continue
		}

		if handler, ok := handlers[route]; ok {
			proxy.routes[route].Patch(environment, handler)
		}
	}
}

func (proxy *ByProxy) Mappings() map[string]string {
	mappings := make(map[string]string, len(proxy.routes))
	for name, route := range proxy.routes {
		mappings[name] = route.Environment
	}
	return mappings
}

func (proxy *ByProxy) Status() *Status {
	return &Status{
		Proxies:      proxy.proxyNames,
		Environments: proxy.environmentNames,
		Mappings:     proxy.Mappings(),
	}
}

func initRoutes(proxies []*ProxyConfig) ([]string, routeMap, error) {

	routeMap := make(routeMap, len(proxies))
	proxyNames := make([]string, len(proxies))

	for i, proxyConfig := range proxies {

		if _, exists := routeMap[proxyConfig.Name]; exists {
			return nil, nil, fmt.Errorf("duplicate proxy name: %s", proxyConfig.Name)
		}

		routeMap[proxyConfig.Name] = NewRoute()
		proxyNames[i] = proxyConfig.Name
	}

	return proxyNames, routeMap, nil
}

func initServers(servers []*ServerConfig) (serverMap, error) {

	serverMap := make(serverMap, len(servers))

	for _, server := range servers {

		if _, exists := serverMap[server.Name]; exists {
			return nil, fmt.Errorf("duplicate server name: %s", server.Name)
		}

		serverMap[server.Name] = server
	}

	return serverMap, nil
}

func initEnvironments(environments []*EnvironmentConfig, routes routeMap, servers serverMap) ([]string, environmentMap, error) {

	environmentMap := make(environmentMap, len(environments))
	environmentNames := make([]string, len(environments))

	for i, environment := range environments {

		if _, exists := environmentMap[environment.Name]; exists {
			return nil, nil, fmt.Errorf("duplicate environment name: %s", environment.Name)
		}

		handlerMap := make(handlerMap, len(environment.Mappings))
		environmentMap[environment.Name] = handlerMap
		environmentNames[i] = environment.Name

		for proxy, server := range environment.Mappings {

			if _, ok := routes[proxy]; !ok {
				return nil, nil, fmt.Errorf("no proxy: %s for mapping in environment: %s", proxy, environment.Name)
			}

			serverConfig, ok := servers[server]
			if !ok {
				return nil, nil, fmt.Errorf("no server: %s for mapping in environment: %s", server, environment.Name)
			}

			baseUrl, e := url.Parse(serverConfig.BaseUrl)
			if e != nil {
				return nil, nil, e
			}

			handler := NewFilterProxy(baseUrl, nil)
			handlerMap[proxy] = handler
		}
	}

	return environmentNames, environmentMap, nil
}

func initServeMux(config *Config, routes routeMap, environments environmentMap) *http.ServeMux {

	serveMux := http.NewServeMux()

	for _, proxy := range config.Proxies {
		serveMux.Handle(proxy.Hostname+"/", routes[proxy.Name])
	}

	for _, environment := range config.Environments {
		for proxy := range environment.Mappings {

			handlerBasename := proxy + "-" + environment.Name
			handlerHostname := handlerBasename + "." + config.Global.Hostname + "/"

			if _, ok := routes[handlerBasename]; !ok {
				serveMux.Handle(handlerHostname, environments[environment.Name][proxy])
			}
		}
	}

	return serveMux
}
