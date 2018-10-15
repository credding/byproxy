package server

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type FilterProxy struct {
	reverseProxy *httputil.ReverseProxy
	filters      []*RequestFilter
}

type RequestFilter interface {
	DoFilter(request *http.Request, next func(*http.Request) (*http.Response)) (*http.Response)
}

type RequestFilterFunc func(request *http.Request, next func(*http.Request) (*http.Response)) (*http.Response)

func (requestFilter RequestFilterFunc) DoFilter(request *http.Request, next func(*http.Request) (*http.Response)) (*http.Response) {
	return requestFilter(request, next)
}

type contextKeyType int

const requestStateKey = contextKeyType(1)

type requestState struct {
	filterSignal    chan int
	responseChannel chan *http.Response
}

func NewFilterProxy(target *url.URL, filters []*RequestFilter) *FilterProxy {

	reverseProxy := httputil.NewSingleHostReverseProxy(target)

	proxy := &FilterProxy{
		reverseProxy: reverseProxy,
		filters:      filters,
	}

	hostnameDirector := reverseProxy.Director
	reverseProxy.Director = func(request *http.Request) {
		state := getState(request)
		go func() {
			hostnameDirector(request)
			request.Host = target.Host
			proxy.doFilter(request, proxy.filters)
			close(state.filterSignal)
		}()
		<-state.filterSignal
	}

	reverseProxy.ModifyResponse = func(response *http.Response) error {
		state := getState(response.Request)
		state.responseChannel <- response
		<-state.filterSignal
		return nil
	}

	reverseProxy.ErrorHandler = func(rw http.ResponseWriter, request *http.Request, err error) {
		state := getState(request)
		close(state.responseChannel)
		<-state.filterSignal
		log.Print(err)
		rw.WriteHeader(http.StatusBadGateway)
	}

	return proxy
}

func (proxy *FilterProxy) ServeHTTP(rw http.ResponseWriter, request *http.Request) {

	state := &requestState{
		filterSignal:    make(chan int),
		responseChannel: make(chan *http.Response),
	}

	ctx := context.WithValue(request.Context(), requestStateKey, state)
	proxy.reverseProxy.ServeHTTP(rw, request.WithContext(ctx))
}

func (proxy *FilterProxy) doFilter(request *http.Request, filters []*RequestFilter) (*http.Response) {

	if len(filters) > 0 {

		filter := *filters[0]
		return filter.DoFilter(request, func(request *http.Request) (*http.Response) {
			return proxy.doFilter(request, filters[1:])
		})

	} else {
		state := getState(request)
		state.filterSignal <- 0
		return <-state.responseChannel
	}
}

func getState(request *http.Request) *requestState {
	return request.Context().Value(requestStateKey).(*requestState)
}
