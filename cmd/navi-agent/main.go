/*
                      _ooOoo_
                     o8888888o
                     88" . "88
                     (| -_- |)
                     O\  =  /O
                  ____/`---'\____
                .'  \\|     |//  `.
               /  \\|||  :  |||//  \
              /  _||||| -:- |||||-  \
              |   | \\\  -  /// |   |
              | \_|  ''\---/''  |   |
              \  .-\__  `-`  ___/-. /
            ___`. .'  /--.--\  `. . __
         ."" '<  `.___\_<|>_/___.'  >'"".
        | | :  `- \`.;`\ _ /`;.`/ - ` : | |
        \  \ `-.   \_ __\ /__ _/   .-` /  /
   ======`-.____`-.___\_____/___.-`____.-'======
                      `=---='
   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
				佛祖保佑       永无BUG
*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	metrics "github.com/rcrowley/go-metrics"
	"io/ioutil"
	"kuaishangtong/common/utils/daemon"
	"kuaishangtong/common/utils/log"
	"kuaishangtong/navi/agent"
	"kuaishangtong/navi/registry"
	"strings"
	"time"
)

func main() {
	if !initializeFlags() {
		return
	}

	err := initializeConfig(_flags.Config)
	if err != nil {
		log.Fatal(err)
	}

	if _flags.Daemon {
		daemon.SetWorkerLogPath(defaultConfig.Log.File)
		daemon.SetLogPath(defaultConfig.Log.File + ".monitor")
		daemon.Exec(daemon.Daemon | daemon.Monitor)
	}

	// log 设置
	logSet := defaultConfig.Log
	log.SetLogFuncCall(logSet.ShowLines)
	log.SetColor(logSet.Coloured)
	log.SetLevel(logSet.Level)
	if _flags.Daemon || logSet.Enable {
		log.SetLogFile(
			logSet.File,
			logSet.Level,
			logSet.Daily,
			logSet.Coloured,
			logSet.ShowLines,
			logSet.MaxDays)
	}

	serverhosts := strings.Split(defaultConfig.Server.ServerHosts, ";")

	serverCount := len(serverhosts)

	var agents []*agent.Agent = make([]*agent.Agent, serverCount, serverCount)

	//zkServers, err := env.GetZookeeperHosts()
	//if err != nil {
	//	zkServers = strings.Split(defaultConfig.Zookeeper.ZookeeperHosts, ";")
	//}
	zkServers := strings.Split(defaultConfig.Zookeeper.ZookeeperHosts, ";")

	for i := 0; i < serverCount; i++ {
		agents[i], err = agent.NewAgent(defaultConfig.Server.ServerName, serverhosts[i], defaultConfig.Server.ServerType)
		if err != nil {
			log.Fatal(err)
		}

		go func(index int) {
			var basePath string

			if defaultConfig.Server.ServerType == "rpc" {
				basePath = defaultConfig.Zookeeper.ZookeeperRPCServicePath
			} else if defaultConfig.Server.ServerType == "http" {
				basePath = defaultConfig.Zookeeper.ZookeeperHTTPServicePath
			}

			r := &registry.ZooKeeperRegister{
				ServiceAddress:   serverhosts[index],
				ZooKeeperServers: zkServers,
				BasePath:         basePath,
				Metrics:          metrics.NewRegistry(),
				UpdateInterval:   2 * time.Second,
			}

			err = r.Start()
			if err != nil {
				log.Fatal(err)
			}

			agents[index].Plugins.Add(r)

			err = agents[index].Serve()
			if err != nil {
				log.Fatal(err)
			}
		}(i)
	}

	type apiUrl struct {
		ApiUrls []string `json:"api_urls"`
	}

	// 为 http 服务添加url注册信息
	if defaultConfig.Server.ServerType == "http" {
		data, err := ioutil.ReadFile(defaultConfig.Server.ServerHttpApiJsonFile)
		if err != nil {
			log.Fatal(err)
		}

		var __apiUrl apiUrl

		err = json.Unmarshal(data, &__apiUrl)
		if err != nil {
			log.Fatal(err)
		}

		urlRegistry, err := libkv.NewStore(store.ZK, zkServers, nil)
		if err != nil {
			log.Fatal(err)
		}

		for _, url := range __apiUrl.ApiUrls {
			key := fmt.Sprintf("%s/%s/%s/%s", strings.Trim(defaultConfig.Zookeeper.ZookeeperURLServicePath, "/"),
				defaultConfig.Server.ServerName, defaultConfig.Server.ServerMode, url)
			err = urlRegistry.Put(key, nil, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
		urlRegistry.Close()
	}

	select {}
}
