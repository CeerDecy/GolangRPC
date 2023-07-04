package config

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github/CeerDecy/RpcFrameWork/crpc/crpcLogger"
	"os"
)

var Conf = &CRConfig{
	logger: crpcLogger.TextLogger(),
}

func init() {
	loadToml()
}

func loadToml() {
	configFile := flag.String("conf", "conf/app.toml", "app config file")
	flag.Parse()
	if _, err := os.Stat(*configFile); err != nil {
		Conf.logger.Debug("config", "conf/app.toml file not load,because not exist")
		return
	}
	_, err := toml.DecodeFile(*configFile, Conf)
	if err != nil {
		Conf.logger.Error("config", "conf/app.toml decode fail check format")
		panic(err)
	}
}

// CRConfig crpc的配置文件
type CRConfig struct {
	logger *crpcLogger.Logger
	Log    map[string]any
	Pool   map[string]any
}
