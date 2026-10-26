package main

import (
	"flag"

	"gopkg.in/ini.v1"
)

type Config struct {
	DbType     string
	DbConn     string
	ListenAddr string
}

func LoadConfig() (*Config, error) {
	file := flag.String("config", "/etc/powerrest.conf", "Configuration file path")

	flag.Parse() // parse command line flags

	return loadConfigFile(*file)
}

// loadConfigFile reads the INI style configuration file at path.
func loadConfigFile(path string) (*Config, error) {
	config := &Config{}

	confFile, err := ini.Load(path)
	if err != nil {
		return config, err
	}

	sect := confFile.Section("")
	config.DbType = sect.Key("db-type").MustString("mysql")
	config.DbConn = sect.Key("db-connection").MustString("")
	config.ListenAddr = sect.Key("listen-address").MustString(":8080")

	return config, nil
}
