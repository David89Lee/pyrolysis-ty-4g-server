package common

import (
	"fmt"
	"log"

	"gopkg.in/ini.v1"
)

type Server struct {
	HttpPort string
	TcpPort  string
}

var ServerSetting = &Server{}

type Log struct {
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

var LogSetting = &Log{}

type Database struct {
	Type     string
	User     string
	Password string
	Host     string
	Name     string
}

var DatabaseSetting = &Database{}

type Dtu struct {
	Dtu1 string `ini:"dtu1"`
	Dtu2 string `ini:"dtu2"`
	Dtu3 string `ini:"dtu3"`
	Dtu4 string `ini:"dtu4"`
}

var DtuSetting = &Dtu{}

var commonCfg *ini.File

func SetUpSetting(CfgFileDir string) {
	var err error

	commonCfgFile := fmt.Sprintf("%v/common.ini", CfgFileDir)

	commonCfg, err = ini.Load(commonCfgFile)
	if err != nil {
		log.Fatalf("setting.Setup, fail to parse 'conf/common.ini': %v", err)
	}

	mapTo(commonCfg, "server", ServerSetting)
	mapTo(commonCfg, "log", LogSetting)
	mapTo(commonCfg, "database", DatabaseSetting)
	mapTo(commonCfg, "dtu", DtuSetting)
}

func mapTo(cfg *ini.File, section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo %s err: %v", section, err)
	}
}
