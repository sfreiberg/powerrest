package main

import (
	"github.com/jteeuwen/ini"

	"flag"
)

type Config struct {
	DbType     string
	DbConn     string
	ListenAddr string
}

func LoadConfig() (*Config, error) {
	file := flag.String("config", "/etc/powerrest.conf", "Configuration file path")
	config := &Config{}
	confFile := ini.New()

	flag.Parse() // parse command line flags

	err := confFile.Load(*file)
	if err != nil {
		return config, err
	}

	sect := confFile.Section("")
	config.DbType = sect.S("db-type", "mysql")
	config.DbConn = sect.S("db-connection", "")
	config.ListenAddr = sect.S("listen-address", ":8080")

	return config, nil
}
