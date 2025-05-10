package config

type Config struct {
	GRPC struct {
		Port int `yaml:"port" env-default:"50051"`
	} `yaml:"grpc"`
}