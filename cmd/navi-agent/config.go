package main

import (
	"bufio"
	"bytes"
	"fmt"
	"git.oschina.net/kuaishangtong/common/utils/json"
	"io"
	"io/ioutil"
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

type zookeeprConf struct {
	ZookeeperHosts           []string `json:"zookeeper_hosts"`
	ZookeeperRPCServicePath  string   `json:"zookeeper_rpc_service_path"`
	ZookeeperHTTPServicePath string   `json:"zookeeper_http_service_path"`
	ZookeeperURLServicePath  string   `json:"zookeeper_url_service_path"`
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

type serverConf struct {
	ServerType  string   `json:"server_type"`
	ServerName  string   `json:"server_name"`
	ServerHosts []string `json:"server_hosts"`
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
