package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env:"ENV" env-default:"local" env-required:"true"`
	StoragePath string `yaml:"storage_path" env:"STORAGE_PATH" env-default:"./storage/storage.db" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"ADDRESS" env-default:"localhost:8082" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"4s" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" env-default:"60s" env-required:"true"`
	User        string        `yaml:"user" env:"HTTP_SERVER_USER" env-default:"" env-required:"true"`
	Password    string        `yaml:"password" env:"HTTP_SERVER_PASSWORD" env-default:"" env-required:"true"`
}

func MustLoad() *Config {
	configPath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok {
		log.Fatal("CONFIG_PATH environment variable is not set")
	} else {
		log.Printf("CONFIG_PATH: %s", configPath)
	}
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set!!!")
	}

	//check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %s", err)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("error reading config file: %s", err)
	}

	return &cfg
}
