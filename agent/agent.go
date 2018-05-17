package agent

import (
	"fmt"
	"kuaishangtong/common/utils/log"
	"strings"
	"time"
)

type Agent struct {
	Plugins           PluginContainer
	agenter           Agenter
	serverName        string
	address           string
	typ               string
	isDocker          bool
	restartServerFunc func() error
}

// NewServer returns a server.
func NewAgent(server_name, address string, typ string, is_docker bool, restartFunc func() error) (*Agent, error) {
	var err error

	if is_docker {
		fields := strings.Split(address, ":")
		if len(fields) != 2 {
			return nil, fmt.Errorf("address %s have no port", address)
		}
		address = fmt.Sprintf("127.0.0.1:%s", fields[1])
	}

	a := &Agent{
		Plugins:           &pluginContainer{},
		serverName:        server_name,
		address:           address,
		typ:               typ,
		isDocker:          is_docker,
		restartServerFunc: restartFunc,
	}

	switch typ {
	case "rpc":
		a.agenter, err = a.NewThriftAgenter()
		if err != nil {
			return nil, err
		}

	case "http":
		a.agenter, err = a.NewHttpAgenter()
		if err != nil {
			return nil, err
		}

	case "grpc":
		a.agenter, err = a.NewGrpcAgenter()
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown server type")
	}

	return a, nil

}

// Serve starts and listens RPC requests.
func (a *Agent) Serve() (err error) {
	var service_active bool = true
	serviceMode, err := a.agenter.ServiceMode()
	if err != nil {
		log.Error(err)
		service_active = false
		goto LOOP
	}

	serviceMode = strings.Trim(serviceMode, "\"")

	log.Infof("service %s mode: [%s]", a.serverName, serviceMode)

	_, err = a.agenter.Ping()
	if err != nil {
		log.Error(err)
		service_active = false
		goto LOOP
	}

	err = a.RegisterName(a.serverName, serviceMode, nil, a.serverName)
	if err != nil {
		log.Error(err)
		service_active = false
		goto LOOP
	}

	log.Infof("register service %s successful", a.serverName)

LOOP:
	pingTicker := time.NewTicker(3 * time.Second)
	defer pingTicker.Stop()

	serviceTicker := time.NewTicker(time.Minute)
	defer serviceTicker.Stop()

	for {

		if a.agenter != nil {
			a.agenter.Close()
		}

		switch a.typ {
		case "rpc":
			a.agenter, err = a.NewThriftAgenter()

		case "http":
			a.agenter, err = a.NewHttpAgenter()

		default:
			err = fmt.Errorf("unknown service type")
		}

		if err != nil {
			log.Error(err)
			a.agenter = nil
		}

		select {
		case <-pingTicker.C:
			if a.agenter != nil {

				_serviceMode, err := a.agenter.ServiceMode()
				if err == nil {
					serviceMode = _serviceMode
				}

				_, err = a.agenter.Ping()
				if err != nil {
					if service_active {
						log.Infof("unregister %s service %s successful", a.typ, a.serverName)
						service_active = false
						err = a.UnRegisterName(a.serverName, serviceMode)
						if err != nil {
							log.Error(err)
							continue
						}
						log.Infof("unregister %s service %s successful", a.typ, a.serverName)
					}
					continue
				}

				if !service_active {
					err = a.RegisterName(a.serverName, serviceMode, nil, a.serverName)
					if err != nil {
						log.Error(err)
						continue
					} else {
						service_active = true
						log.Infof("register %s service %s successful", a.typ, a.serverName)
					}
				}
			}
		case <-serviceTicker.C:
		}
	}

	return nil
}

func (a *Agent) RegisterName(name, mode string, rcvr interface{}, metadata string) error {
	if a.Plugins == nil {
		a.Plugins = &pluginContainer{}
	}

	return a.Plugins.DoRegister(name+"/"+mode, rcvr, metadata)
}

func (a *Agent) UnRegisterName(name, mode string) error {
	if a.Plugins == nil {
		a.Plugins = &pluginContainer{}
	}

	return a.Plugins.DoUnRegister(name + "/" + mode)
}
