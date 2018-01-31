package navicli

type ConnPool interface {
	GetConn() (interface{}, error)
	Close()
}
