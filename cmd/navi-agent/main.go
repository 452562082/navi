package main

import (
	"git.oschina.net/kuaishangtong/common/utils/daemon"
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/agent"
	"git.oschina.net/kuaishangtong/navi/registry"
	metrics "github.com/rcrowley/go-metrics"
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

	var agents []*agent.Agent = make([]*agent.Agent, len(defaultConfig.Server.ServerHosts), len(defaultConfig.Server.ServerHosts))
	for i := 0; i < len(defaultConfig.Server.ServerHosts); i++ {
		agents[i], err = agent.NewAgent(defaultConfig.Server.ServerName, defaultConfig.Server.ServerHosts[i], defaultConfig.Server.ServerType)
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
				ServiceAddress:   defaultConfig.Server.ServerHosts[index],
				ZooKeeperServers: defaultConfig.Zookeeper.ZookeeperHosts,
				BasePath:         basePath,
				Metrics:          metrics.NewRegistry(),
			}

			err = r.Start()
			if err != nil {
				log.Fatal(err)
			}

			agents[i].Plugins.Add(r)

			err = agents[i].Serve()
			if err != nil {
				log.Fatal(err)
			}
		}(i)
	}

	// 为 http 服务添加url注册信息
	if defaultConfig.Server.ServerType == "http" {

	}

	select {}
}
