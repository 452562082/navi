package navicli

type ConnPool interface {
	GetConn() (interface{}, error)
	PutConn(conn interface{}) error
	SetServerHostUnavailable(interface{})
	//ClearInvalidHost()
	GetFailMode() (interface{})
	GetRetries() (int)
	Close()
}