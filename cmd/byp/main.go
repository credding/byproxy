package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/credding/byproxy/pkg/cli"
	"github.com/credding/byproxy/pkg/server"
	"github.com/credding/byproxy/pkg/util"
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

	config, e := util.LoadConfig(configLocation)
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
		}
		if status != nil {
			printStatus(status)
		}
	}
}

func startServer(config *util.Config) error {

	fmt.Println("Starting ByProxy server")

	proxy, e := server.NewByProxy(config)
	if e != nil {
		return e
	}

	if mappings, e := util.LoadMappings(); mappings != nil {
		proxy.LoadMappings(mappings)
	} else if e != nil {
		fmt.Println(e)
	}

	if e := proxy.ListenAndServe(); e != nil {
		return e
	}
	return nil
}

func clientCommand(config *util.Config, command string, args []string) (*server.Status, error) {

	global := config.Global
	addr := fmt.Sprintf("%s:%d", global.Addr, global.Port)
	remote, e := cli.NewRemote(addr)
	if e != nil {
		return nil, e
	}

	switch command {
	case "reload":
		fmt.Println("Reloading ByProxy configuration")
		return remote.Reload()
	case "status":
		return remote.Status()
	case "use":
		return remote.Use(args[1], args[2:]...)
	case "stop":
		fmt.Println("Stopping ByProxy server")
		return nil, remote.Stop()
	default:
		return nil, errors.New("Unknown command: " + command)
	}
}

func printStatus(status *server.Status) {
	fmt.Println("Mappings:")
	for proxy, mapping := range status.Mappings {
		fmt.Println(proxy + " -> " + mapping)
	}
	fmt.Println("Environments:")
	for _, environment := range status.Environments {
		fmt.Println("- " + environment)
	}
}
