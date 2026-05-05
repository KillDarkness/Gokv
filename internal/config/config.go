package config

import "fmt"

type Config struct {
	Host       string
	Port       int
	Databases  int
	AppendOnly bool
	Snapshot   bool
}

func Default() Config {
	return Config{
		Host:       "0.0.0.0",
		Port:       6379,
		Databases:  1,
		AppendOnly: false,
		Snapshot:   false,
	}
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
