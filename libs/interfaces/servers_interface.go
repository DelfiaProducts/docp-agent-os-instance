package interfaces

// IServerInterface is interface for servers
type IServerInterface interface {
	Setup() error
}

// IServerRegisterInterface is interface for servers
type IServerRegisterInterface interface {
	Setup() error
	SendMetadata(data []byte) ([]byte, error)
}
