package websocket

// New creates a new object and return interface
func New() Interface {
	var wst = new(impl)
	return wst
}
