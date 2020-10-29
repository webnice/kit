// +build go1.8

package wrapsrw

import "net/http"

// Go 1.8 http Push

func (f *http2FancyWriter) Push(target string, opts *http.PushOptions) error {
	return f.basic.ResponseWriter.(http.Pusher).Push(target, opts)
}
