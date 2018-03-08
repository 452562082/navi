package navicli

type ConnPool interface {
	GetConn() (interface{}, error)
	SetServerConnPoolUnavailable(interface{})
	//ClearInvalidHost()
	GetFailMode() (interface{})
	GetRetries() (int)
	Close()
}