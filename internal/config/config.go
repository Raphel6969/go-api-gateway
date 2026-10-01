package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port string `yaml:"port"`
}

type AuthConfig struct {
	JWTSecret string            `yaml:"jwt_secret"`
	APIKeys   map[string]string `yaml:"api_keys"`
}

type RateLimitConfig struct {
	Capacity   float64 `yaml:"capacity"`
	RefillRate float64 `yaml:"refill_rate"`
}

type CacheConfig struct {
	TTLSeconds int `yaml:"ttl_seconds"`
}

type Route struct {
	Path    string   `yaml:"path"`
	Targets []string `yaml:"targets"`
	Public  bool     `yaml:"public"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	Enabled  bool   `yaml:"enabled"`
}

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Auth      AuthConfig      `yaml:"auth"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Cache     CacheConfig     `yaml:"cache"`
	Redis     RedisConfig     `yaml:"redis"`
	Routes    []Route         `yaml:"routes"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
