package config

import (
	"log"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DatacenterID *int    `yaml:"datacenter_id" env-required:"true"`
	MachineID    *int    `yaml:"machine_id" env-required:"true"`
	IsDebug      *bool   `yaml:"is_debug" env-required:"true"`
	Listen       *Listen `yaml:"listen"`
	Log          *Log    `yaml:"log"`
}

type Listen struct {
	Host *string `yaml:"host" env-required:"true"`
	Port *string `yaml:"port" env-required:"true"`
	Grpc *string `yaml:"grpc" env-required:"true"`
}

type Log struct {
	Path     *string `yaml:"path" env-required:"true"`
	Filename *string `yaml:"filename" env-required:"true"`
}

var instance *Config
var once sync.Once

func Init() *Config {
	once.Do(func() {

		pwd, err := os.Getwd()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		pathConfig := pwd + "/config.yml"

		log.Println("read application configuration: pwd: ", pwd)

		instance = &Config{}
		if err := cleanenv.ReadConfig(pathConfig, instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Println(help)
			log.Fatal(err)
		}
	})
	return instance
}
