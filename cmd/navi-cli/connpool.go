package navicli

import "git.oschina.net/kuaishangtong/navi/lb"

type ConnPool interface {
	GetConn() (interface{}, error)
	AddInvalidHost(interface{})
	ClearInvalidHost()
	GetFailMode() (lb.FailMode)
	GetRetries() (int)
	Close()
}