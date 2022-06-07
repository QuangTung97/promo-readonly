package config

import "fmt"

// MemcacheConfig ...
type MemcacheConfig struct {
	Host     string `mapstructure:"host"`
	Port     uint16 `mapstructure:"port"`
	NumConns int    `mapstructure:"num_conns"`
}

// Addr ...
func (c MemcacheConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
