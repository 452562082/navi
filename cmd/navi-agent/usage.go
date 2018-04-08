package main

import (
	"flag"
	"fmt"
	"os"

	"kuaishangtong/common/utils/log"
)

const (
	APP_NAME            = "navi-agent"
	APP_DESCRIPTION     = "navi-agent for rpcserver or httpserver about register service host"
	_APP_VERSION        = "1.0.0"
	__CONF_DEFAULT_PATH = "/usr/local/navi-agent/etc/cfg.json"
)

var _version = "1.0.0"

var _flags AppFlags

func Version() string {

	if _version != "" {
		return _version
	}
	return _APP_VERSION
}

type AppFlags struct {
	Daemon  bool
	Version bool
	Help    bool
	Config  string
}

func usage() {
	fmt.Printf("%s version: %s, %s\n", APP_NAME, Version(), APP_DESCRIPTION)
	if _flags.Help {
		fmt.Printf("\nusage:\n")
		flag.PrintDefaults()
	}
}

func initializeFlags() bool {

	flag.BoolVar(&_flags.Version, "v", false, "print version")
	flag.BoolVar(&_flags.Daemon, "d", false, "run in daemon")
	flag.BoolVar(&_flags.Help, "h", false, "print this message")
	flag.StringVar(&_flags.Config, "c", __CONF_DEFAULT_PATH, "specify config file")

	flag.Usage = usage
	flag.Parse()

	if _flags.Version || _flags.Help {
		usage()
		return false
	}

	if !exist(_flags.Config) {
		log.Errorf("can not find config file: %s", _flags.Config)
		return false
	}

	return true
}

func exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
