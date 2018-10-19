package main

import (
	"flag"
	"fmt"
	"github.com/credding/byproxy/pkg/byproxy"
	"log"
	"os"
)

func main() {

	var configLocation string
	var addr string
	var port int

	mainFlags := flag.NewFlagSet("byp", flag.ExitOnError)
	mainFlags.StringVar(&configLocation, "config", "", "ByProxy config location")
	mainFlags.StringVar(&addr, "addr", "", "ByProxy server address")
	mainFlags.IntVar(&port, "port", 0, "ByProxy server port")

	mainFlags.Parse(os.Args[1:])

	config, e := byproxy.LoadConfig(configLocation)
	if e != nil {
		log.Fatalln(e)
	}

	if addr != "" {
		config.Global.Addr = addr
	}
	if port != 0 {
		config.Global.Port = port
	}

	command := mainFlags.Arg(0)

	if command == "start" {
		e := startServer(config)
		if e != nil {
			fmt.Println(e)
		}

	} else {
		status, e := clientCommand(config, command, mainFlags.Args())
		if e != nil {
			fmt.Println(e)
		} else if status != nil {
			printStatus(status)
		}
	}
}

func startServer(config *byproxy.Config) error {
	if server, e := byproxy.NewServer(config); e != nil {
		return e
	} else {
		log.Println("starting ByProxy server")
		printStatus(server.Status())
		return server.ListenAndServe()
	}
}

func clientCommand(config *byproxy.Config, command string, args []string) (*byproxy.Status, error) {

	address := fmt.Sprintf("%s:%d", config.Global.Addr, config.Global.Port)
	remote, e := byproxy.NewRemote(address)
	if e != nil {
		return nil, e
	}

	switch command {
	case "reload":
		fmt.Println("reloading ByProxy configuration")
		return remote.Reload(config)
	case "status":
		return remote.Status()
	case "use":
		return remote.Use(args[1], args[2:]...)
	case "stop":
		fmt.Println("stopping ByProxy server")
		return nil, remote.Stop()
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

func printStatus(status *byproxy.Status) {
	fmt.Println("mappings:")
	for _, proxy := range status.Proxies {
		fmt.Printf(" %s -> %s\n", proxy, status.Mappings[proxy])
	}
	fmt.Println("environments:")
	for _, environment := range status.Environments {
		fmt.Println(" - " + environment)
	}
}
