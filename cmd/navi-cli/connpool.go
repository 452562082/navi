package navicli

type ConnPool interface {
	GetConn() (interface{}, error)
	AddInvalidHost(string)
	ClearInvalidHost()
	Close()
}
