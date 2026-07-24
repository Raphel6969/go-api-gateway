package config

type Route struct {
	Path    string
	Targets []string
}

type Config struct {
	Port   string
	Routes []Route
}

func LoadStaticConfig() *Config {
	return &Config{
		Port: ":8080",
		Routes: []Route{
			{
				Path: "/users",
				Targets: []string{
					"http://localhost:8081",
					"http://localhost:8082",
				},
			},
		},
	}
}
