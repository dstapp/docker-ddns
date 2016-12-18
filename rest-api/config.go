package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	SharedSecret string
	Server string
	Zone string
	Domain string
	NsupdateBinary string
	RecordTTL int
}

func (conf *Config) LoadConfig(path string) {
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
