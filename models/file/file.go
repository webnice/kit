package file // import "github.com/webnice/kit/v1/models/file"

// New creates new object and return Interface
func New() Interface {
	var obj = new(impl)
	return obj
}
