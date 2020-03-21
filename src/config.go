package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ListenPort int
	ListenAddress string

	KubeConfigPath string
}

func NewConfig() *Config {
	return &Config{
		ListenPort:    getIntOrDefault("LISTEN_PORT", 8888),
		ListenAddress: getStringOrDefault("LISTEN_ADDR", "0.0.0.0"),
		KubeConfigPath: getStringOrDefault("KUBECONFIG", fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))),
	}
}

func (c *Config) AddressAndPort() string {
	return fmt.Sprintf("%s:%d", c.ListenAddress, c.ListenPort)
}

func getStringOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntOrDefault(key string, defaultValue int) int {
	s := os.Getenv(key)
	if s == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(s)
	if err != nil {
		panic(errors.New(fmt.Sprintf("environment variable (%s) value %s is not an integer", key, s)))
	}
	return value
}

