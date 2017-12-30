package file // import "gopkg.in/webnice/kit.v1/models/file"

//import "gopkg.in/webnice/log.v2"
//import "gopkg.in/webnice/debug.v1"
import ()

// New creates new object and return Interface
func New() Interface {
	var obj = new(impl)
	return obj
}
