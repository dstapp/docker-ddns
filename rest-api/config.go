package main

import (
	"encoding/json"
	"flag"
	"os"
)

type Config struct {
	SharedSecret   string
	Server         string
	Zone           string
	Domain         string
	NsupdateBinary string
	RecordTTL      int
	Port           int
}

type ConfigFlags struct {
	Config
	ConfigFile      string
	DoNotLoadConfig bool
	LogLevel        int
}

func (conf *Config) loadConfigFromFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		panic(err)
	}
}

func (flagsConf *ConfigFlags) setupFlags() {
	flag.BoolVar(&flagsConf.DoNotLoadConfig, "noConfig", false, "Do not load the config file")
	flag.StringVar(&flagsConf.ConfigFile, "c", "/etc/dyndns.json", "The configuration file")
	flag.StringVar(&flagsConf.SharedSecret, "sharedSecret", "", "The shared secret (default a generated random string)")
	flag.StringVar(&flagsConf.Server, "server", "localhost", "The address of the bind server")
	flag.StringVar(&flagsConf.Zone, "zone", "localhost", "Zone")
	flag.StringVar(&flagsConf.Domain, "domain", "localhost", "Domain")
	flag.StringVar(&flagsConf.NsupdateBinary, "nsupdateBinary", "nsupdate", "Path to nsupdate program")
	flag.IntVar(&flagsConf.RecordTTL, "recordTTL", 300, "RecordTTL")
	flag.IntVar(&flagsConf.Port, "p", 8080, "Port")
	flag.IntVar(&flagsConf.LogLevel, "log", 0, "Set the log level")
}

// LoadConfig loads config values from the config file and from the passed arguments.
// Gives command line arguments precedence.
func (flagsConf *ConfigFlags) LoadConfig() {
	flagsConf.setupFlags()
	flag.Parse()

	if !flagsConf.DoNotLoadConfig {
		flagsConf.loadConfigFromFile(flagsConf.ConfigFile)
		flag.Parse() // Parse a second time to overwrite settings from the loaded file
	}

	// Fix unsafe config values
	if flagsConf.SharedSecret == "" {
		flagsConf.SharedSecret = randomString()
	}
}
