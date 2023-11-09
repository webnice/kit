package server

// InterfaceGrpc Интерфейс пакета.
type InterfaceGrpc interface {
}

// Объект сущности, реализующий интерфейс InterfaceServerGrpc.
type implGrpc struct {
	parent *impl
}
