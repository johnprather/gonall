package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"

	yaml "gopkg.in/yaml.v2"
)

const configFileName = ".gonall.yml"

// Config struct defines layout of configuration file (app tunables)
type Config []ConfigBlock

// ConfigBlock is the configurables for a matched set of hosts
type ConfigBlock struct {
	Host      string        `yaml:"Host"`
	User      string        `yaml:"User"`
	Port      int           `yaml:"Port"`
	Timeout   time.Duration `yaml:"Timeout"`
	ProxyHost Server        `yaml:"ProxyHost"`
}

var config Config

func (c Config) load() {
	home := os.Getenv("HOME")

	// if no home, assume no config
	if home == "" {
		return
	}

	configFile := path.Join(home, configFileName)

	// if no config file, assume no config ;)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return
	}

	// having chosen a config file, use it!
	c.read(configFile)
}

func (c *Config) read(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read config file %s: %s\n", filename, err)
		os.Exit(1)
	}
	var newConfig Config
	err = yaml.Unmarshal(data, &newConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to unmarshal yaml into config")
		os.Exit(1)
	}
	config = newConfig
}

// forServer returns a Config block suitable for the specified Server
func (c Config) forServer(server Server) ConfigBlock {
	hostConfig := ConfigBlock{}
	for i := len(config) - 1; i >= 0; i-- {
		if match, err := filepath.Match(config[i].Host, string(server)); err == nil && match {
			if config[i].ProxyHost != "" {
				if config[i].ProxyHost == "none" {
					hostConfig.ProxyHost = ""
				} else if config[i].ProxyHost != server {
					hostConfig.ProxyHost = config[i].ProxyHost
				}
			}
			if config[i].Timeout != 0 {
				hostConfig.Timeout = config[i].Timeout
			}
			if config[i].User != "" {
				hostConfig.User = config[i].User
			}

			if config[i].Port != 0 {
				hostConfig.Port = config[i].Port
			}
		}
	}

	// replace empty values with defaults
	if hostConfig.User == "" {
		hostConfig.User = "root"
	}
	if hostConfig.Port == 0 {
		hostConfig.Port = 22
	}

	// commandline overrides
	if flags.timeout != 0 {
		hostConfig.Timeout = time.Duration(flags.timeout)
	}
	if flags.user != "" {
		hostConfig.User = flags.user
	}

	hostConfig.Host = string(server)

	return hostConfig
}

// hostPort only works if Host has been set to a valid hostname,
// as is the case when obtained with forServer(server)
func (cb ConfigBlock) hostPort() string {
	if match, err := regexp.MatchString(":[0-9]+$", cb.Host); err == nil && match {
		return cb.Host
	}
	return fmt.Sprintf("%s:%d", cb.Host, cb.Port)
}
