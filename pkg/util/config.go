package util

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"syscall"
)

type Config struct {
	Global       *Global
	Proxies      []*Proxy
	Servers      []*Server
	Environments []*Environment
	Filters      []*Filter
}

type Global struct {
	Hostname string
	Addr     string
	Port     int
	Filters  []*FilterMapping
}

type Proxy struct {
	Name     string
	Hostname string
	Filters  []*FilterMapping
}

type Server struct {
	Name    string
	BaseUrl string `yaml:"base-url"`
	Filters []*FilterMapping
}

type Environment struct {
	Name     string
	Mappings map[string]string
	Filters  []*FilterMapping
}

type Filter struct {
	Name    string
	Handler string
	Config  FilterConfig
}

type FilterMapping struct {
	Name   string
	Config FilterConfig
}

type FilterConfig map[string]interface{}

func LoadConfig(location string) (*Config, error) {

	config := &Config{}

	if location == "" {
		location = path.Join(userHomeDir(), ".byp/config.yaml")
	}

	if e := config.loadConfig(location); e != nil {
		return nil, e
	}
	config.setDefaults()

	return config, nil
}

func LoadMappings() (map[string]string, error) {

	location := path.Join(userHomeDir(), ".byp/mappings")
	mappings := make(map[string]string)

	bytes, e := ioutil.ReadFile(location)
	if e != nil {
		if e, ok := e.(*os.PathError); ok {
			if e.Err == syscall.ENOENT {
				return nil, nil
			}
		}
		return nil, e
	}
	if e = yaml.Unmarshal(bytes, &mappings); e != nil {
		return nil, e
	}
	return mappings, nil
}

func WriteMappings(mappings map[string]string) error {

	location := path.Join(userHomeDir(), ".byp/mappings")

	bytes, e := yaml.Marshal(mappings)
	if e != nil {
		return e
	}
	if e = ioutil.WriteFile(location, bytes, 0644); e != nil {
		return e
	}
	return nil
}

func (config *Config) loadConfig(location string) error {

	bytes, e := ioutil.ReadFile(location)
	if e != nil {
		return e
	}
	if e = yaml.Unmarshal(bytes, config); e != nil {
		return e
	}
	return nil
}

func (config *Config) setDefaults() {

	if config.Global == nil {
		config.Global = &Global{"localhost", "127.0.0.1", 8042, nil}
	} else {
		if config.Global.Hostname == "" {
			config.Global.Hostname = "localhost"
		}
		if config.Global.Addr == "" {
			config.Global.Addr = "127.0.0.1"
		}
		if config.Global.Port == 0 {
			config.Global.Port = 8042
		}
	}

	for _, proxy := range config.Proxies {
		if proxy.Hostname == "" {
			proxy.Hostname = proxy.Name + "." + config.Global.Hostname
		}
	}

	usages := make(map[string][]FilterConfig)

	for _, filter := range config.Filters {
		usages[filter.Name] = make([]FilterConfig, 0)
	}

	collectFilterUsages(usages, config.Global.Filters)
	for _, proxy := range config.Proxies {
		collectFilterUsages(usages, proxy.Filters)
	}
	for _, server := range config.Servers {
		collectFilterUsages(usages, server.Filters)
	}
	for _, environment := range config.Environments {
		collectFilterUsages(usages, environment.Filters)
	}

	for _, filter := range config.Filters {
		setFilterDefaults(usages, filter)
	}
}

func collectFilterUsages(usages map[string][]FilterConfig, filters []*FilterMapping) {
	for _, usage := range filters {
		if usage.Config == nil {
			usage.Config = make(FilterConfig)
		}
		usages[usage.Name] = append(usages[usage.Name], usage.Config)
	}
}

func setFilterDefaults(usages map[string][]FilterConfig, filter *Filter) {
	for _, config := range usages[filter.Name] {
		for key, value := range filter.Config {
			if _, exists := config[key]; !exists {
				config[key] = value
			}
		}
	}
}
