package config

// JaegerConfig ...
type JaegerConfig struct {
	Host string `mapstructure:"host"`
	Port uint16 `mapstructure:"port"`
}
