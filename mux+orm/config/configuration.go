package config

type Configuration struct {
	Database DatabaseConfiguration
	Server   ServerConfiguration
}

type DatabaseConfiguration struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

type ServerConfiguration struct {
	Port string
}
