package config

type Route struct {
	Path   string
	Target string
}

type Config struct {
	Port   string
	Routes []Route
}

func LoadStaticConfig() *Config {
	return &Config{
		Port: "8080",
		Routes: []Route{
			{
				Path:   "/users",
				Target: "http://localhost:8081",
			},
			{
				Path:   "/orders",
				Target: "http://localhost:8082",
			},
		},
	}
}
