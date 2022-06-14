package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Address    string `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	Key        string `env:"KEY" `
	DBURI      string `env:"DATABASE_URI" envDefault:"postgres://postgres:postgres@localhost:5432/gophermart"`
	AccSysSddr string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://127.0.0.1:8080"`
}

func LoadConfig() Config {
	logrus.SetReportCaller(true)
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		logrus.Error(err)
	}
	return cfg
}
