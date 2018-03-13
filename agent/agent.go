package agent

import (
	"fmt"
	"kuaishangtong/common/utils/log"
	"strings"
	"time"
)

type Agent struct {
	Plugins    PluginContainer
	agenter    Agenter
	servername string
	address    string
	typ        string
}

// NewServer returns a server.
func NewAgent(servername, address string, typ string, is_docker bool) (*Agent, error) {
	var err error

	fields := strings.Split(address, ":")
	if len(fields) != 2 {
		return nil, fmt.Errorf("address %s have no port", address)
	}
	address = fmt.Sprintf("127.0.0.1:%s", fields[1])

	a := &Agent{
		Plugins:    &pluginContainer{},
		servername: servername,
		address:    address,
		typ:        typ,
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

	default:
		return nil, fmt.Errorf("unknown server type")
	}

	return a, nil

}

// Serve starts and listens RPC requests.
func (a *Agent) Serve() (err error) {

	serviceMode, err := a.agenter.ServiceMode()
	if err != nil {
		return err
	}

	_, err = a.agenter.Ping()
	if err != nil {
		return err
	}

	err = a.RegisterName(a.servername, serviceMode, nil, a.servername)
	if err != nil {
		return err
	}

	log.Infof("register service %s successful", a.servername)

	var service_active bool = true

	pingTicker := time.NewTicker(time.Second)
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
						service_active = false
						err = a.UnRegisterName(a.servername, serviceMode)
						if err != nil {
							log.Error(err)
							continue
						}
						log.Infof("unregister %s service %s successful", a.typ, a.servername)
					}
					continue
				}

				if !service_active {
					err = a.RegisterName(a.servername, serviceMode, nil, a.servername)
					if err != nil {
						log.Error(err)
						continue
					} else {
						service_active = true
						log.Infof("register %s service %s successful", a.typ, a.servername)
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
