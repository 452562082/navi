package navicli

import "git.oschina.net/kuaishangtong/navi/lb"

type ConnPool interface {
	GetConn() (interface{}, error)
	AddInvalidHost(string)
	ClearInvalidHost()
	GetFailMode() (lb.FailMode)
	Close()
}