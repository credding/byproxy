package byproxy

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
)

type Config struct {
	Global       *GlobalConfig
	Proxies      []*ProxyConfig
	Servers      []*ServerConfig
	Environments []*EnvironmentConfig
	Filters      []*FilterConfig
}

type GlobalConfig struct {
	Hostname string
	Addr     string
	Port     int
	Filters  []*FilterMapping
}

type ProxyConfig struct {
	Name     string
	Hostname string
	Filters  []*FilterMapping
}

type ServerConfig struct {
	Name    string
	BaseUrl string `yaml:"base-url"`
	Filters []*FilterMapping
}

type EnvironmentConfig struct {
	Name     string
	Mappings map[string]string
	Filters  []*FilterMapping
}

type FilterConfig struct {
	Name    string
	Handler string
	Config  FilterOptions
}

type FilterMapping struct {
	Name   string
	Config FilterOptions
}

type FilterOptions map[string]interface{}

func LoadConfig(location string) (*Config, error) {

	if location == "" {
		location = path.Join(userHomeDir(), ".byp/config.yaml")
	}

	bytes, e := ioutil.ReadFile(location)
	if e != nil {
		return nil, e
	}

	config := &Config{}
	if e = yaml.Unmarshal(bytes, config); e != nil {
		return nil, e
	}

	config.SetDefaults()

	return config, nil
}

func (config *Config) SetDefaults() {

	if config.Global == nil {
		config.Global = &GlobalConfig{"localhost", "127.0.0.1", 8042, nil}
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

	usages := make(map[string][]FilterOptions)

	for _, filter := range config.Filters {
		usages[filter.Name] = make([]FilterOptions, 0)
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
		addImplicitMappings(config, environment)
	}

	for _, filter := range config.Filters {
		setFilterDefaults(usages, filter)
	}
}

func collectFilterUsages(usages map[string][]FilterOptions, filters []*FilterMapping) {
	for _, usage := range filters {
		if usage.Config == nil {
			usage.Config = make(FilterOptions)
		}
		usages[usage.Name] = append(usages[usage.Name], usage.Config)
	}
}

func setFilterDefaults(usages map[string][]FilterOptions, filter *FilterConfig) {
	for _, config := range usages[filter.Name] {
		for key, value := range filter.Config {
			if _, exists := config[key]; !exists {
				config[key] = value
			}
		}
	}
}

func addImplicitMappings(config *Config, environment *EnvironmentConfig) {

	if environment.Mappings == nil {
		environment.Mappings = make(map[string]string)
	}

	for _, proxy := range config.Proxies {
		if environment.Mappings[proxy.Name] != "" {
			continue
		}
		implicitServer := proxy.Name + "-" + environment.Name
		for _, server := range config.Servers {
			if server.Name == implicitServer {
				environment.Mappings[proxy.Name] = server.Name
				break
			}
		}
	}
}
