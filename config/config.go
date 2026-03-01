package config

import (
	"bitcoin-watcher/common"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Telegram Telegram `yaml:"telegram"`
	DB       DB       `yaml:"db"`
	HTTP     HTTP     `yaml:"http"`
	LogLevel string   `yaml:"log_level"`
}

type Telegram struct {
	Enabled  bool   `yaml:"enabled"`
	ChatID   int64  `yaml:"chat_id"`
	BotToken string `yaml:"bot_token"`
}

type DB struct {
	// Host     string `yaml:"host"     binding:"required"`
	// Port     string `yaml:"port"     binding:"required"`
	// User     string `yaml:"user"     binding:"required"`
	// Password string `yaml:"password" binding:"required"`
	// Name     string `yaml:"name"     binding:"required"`
}

type HTTP struct {
	Port int `yaml:"port"     binding:"required,min=1,max=65535"`
}

var Main Config

func InitFrom(path string) {
	body, err := os.ReadFile(path)
	if err != nil {
		common.Logger().With(err).Fatalw("can't read config file in the selected path", "path", path)
	}
	err = yaml.Unmarshal(body, &Main)
	if err != nil {
		common.Logger().With(err).Fatalf("can't unmarshal config file in the selected path")
	}
	if err := binding.Validator.ValidateStruct(Main); err != nil {
		common.Logger().With(err).Fatalw("config validation failed", "error", err)
	}
}
