package navicli

type ConnPool interface {
	GetConn() (interface{}, error)
	PutConn(interface{}) error
	SetServerConnPoolUnavailable(interface{})
	GetFailMode() interface{}
	GetRetries() int
	Close()
}

//type ConnPool interface {
//	GetConn() (RPCConn, error)
//	PutConn(rpcconn RPCConn) error
//	SetServerConnPoolUnavailable(interface{})
//	GetFailMode() interface{}
//	GetRetries() int
//	Close()
//}

type RPCConn interface {
	Close()
	Reconnect()
	Available() bool
	HostName() string
}
