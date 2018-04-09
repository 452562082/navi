package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"kuaishangtong/common/utils/json"
	"kuaishangtong/common/utils/log"
	"strings"
)

var defaultConfig Config //程序主配置

func init() {
	defaultConfig.Version = "V1.0.0"
}

type Config struct {
	// server
	Server serverConf `json:"server"`

	// zk
	Zookeeper zookeeprConf `json:"zookeeper"`

	// log
	Log     logConf `json:"log"`
	Version string  `json:"version"`
}

type serverConf struct {
	ServerType            string `json:"server_type"`
	ServerMode            string `json:"server_mode"`
	ServerHttpApiJsonFile string `json:"server_http_api_json_file"`
	ServerName            string `json:"server_name"`
	ServerHosts           string `json:"server_hosts"`
	ServerRestartScript   string `json:"server_restart_script"`
	IsDocker              bool   `json:"is_docker"`
}

type zookeeprConf struct {
	ZookeeperHosts           string `json:"zookeeper_hosts"`
	ZookeeperRPCServicePath  string `json:"zookeeper_rpc_service_path"`
	ZookeeperHTTPServicePath string `json:"zookeeper_http_service_path"`
	ZookeeperURLServicePath  string `json:"zookeeper_url_service_path"`
}

type logConf struct {
	Enable bool   `json:"enable"`
	File   string `json:"file"`
	Level  int    `json:"level"`
	Async  bool   `json:"async"`

	Coloured  bool `json:"coloured"`
	ShowLines bool `json:"show_lines"`

	// Rotate at line
	MaxLines int `json:"maxlines"`

	// Rotate at size
	MaxSize int `json:"maxsize"`

	// Rotate daily
	Daily   bool `json:"daily"`
	MaxDays int  `json:"maxdays"`
}

func (c *Config) init(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%s:%s", err, filename)
	}

	if err = json.Unmarshal(data, &c); err != nil {
		if serr, ok := err.(*json.SyntaxError); ok {
			line, col := getOffsetPosition(bytes.NewBuffer(data), serr.Offset)
			highlight := getHighLightString(bytes.NewBuffer(data), line, col)
			fmt.Printf("\n%v", err)
			fmt.Printf(":\n:Error at line %d, column %d (file offset %d):\n%s",
				line, col, serr.Offset, highlight)
		}
		return err
	}

	if !strings.EqualFold(c.Server.ServerType, "rpc") && !strings.EqualFold(c.Server.ServerType, "http") {
		return fmt.Errorf("illegal server_type: %s, must be 'rpc' or 'http'", c.Server.ServerType)
	}
	log.Noticef("[agent] server_type: %s", c.Server.ServerType)

	if !strings.EqualFold(c.Server.ServerMode, "dev") && !strings.EqualFold(c.Server.ServerMode, "prod") {
		return fmt.Errorf("illegal server_mode: %s, must be 'prod' or 'dev'", c.Server.ServerMode)
	}
	log.Noticef("[agent] server_mode: %s", c.Server.ServerMode)

	if strings.EqualFold(c.Server.ServerType, "http") && len(c.Server.ServerHttpApiJsonFile) == 0 {
		return fmt.Errorf("server_type is 'http', server_http_api_json_file can not be \"\"")
	}

	if len(c.Server.ServerName) == 0 {
		return fmt.Errorf("illegal server_name, server_name can not be \"\"")
	}
	log.Noticef("[agent] server_name: %s", c.Server.ServerName)

	if len(c.Server.ServerHosts) == 0 {
		return fmt.Errorf("illegal server_hosts, length of server_hosts can not be 0")
	}
	log.Noticef("[agent] server_hosts: %s", c.Server.ServerHosts)

	/* zk */
	if len(c.Zookeeper.ZookeeperHosts) == 0 {
		return fmt.Errorf("zookeeper_hosts can not be \"\"")
	}
	log.Noticef("[agent] zookeeper_hosts: %s", c.Zookeeper.ZookeeperHosts)

	if strings.EqualFold(c.Server.ServerType, "rpc") && len(c.Zookeeper.ZookeeperRPCServicePath) == 0 {
		return fmt.Errorf("server_type is 'rpc', zookeeper_rpc_service_path can not be \"\"")
	}

	if strings.EqualFold(c.Server.ServerType, "http") && len(c.Zookeeper.ZookeeperHTTPServicePath) == 0 {
		return fmt.Errorf("server_type is 'http', zookeeper_http_service_path can not be \"\"")
	}

	if strings.EqualFold(c.Server.ServerType, "http") && len(c.Zookeeper.ZookeeperURLServicePath) == 0 {
		return fmt.Errorf("server_type is 'http', zookeeper_url_service_path can not be \"\"")
	}

	return nil
}

func initializeConfig(filename string) error {
	return defaultConfig.init(filename)
}

func getOffsetPosition(f io.Reader, pos int64) (line, col int) {
	line = 1
	br := bufio.NewReader(f)
	thisLine := new(bytes.Buffer)
	for n := int64(0); n < pos; n++ {
		b, err := br.ReadByte()
		if err != nil {
			break
		}
		if b == '\n' {
			thisLine.Reset()
			line++
			col = 1
		} else {
			col++
			thisLine.WriteByte(b)
		}
	}

	return
}

func getHighLightString(f io.Reader, line int, col int) (highlight string) {
	br := bufio.NewReader(f)
	var thisLine []byte
	var err error
	for i := 1; i <= line; i++ {
		thisLine, _, err = br.ReadLine()
		if err != nil {
			fmt.Println(err)
			return
		}
		if i >= line-2 {
			highlight += fmt.Sprintf("%5d: %s\n", i, string(thisLine))
		}
	}
	highlight += fmt.Sprintf("%s^\n", strings.Repeat(" ", col+5))
	return
}
