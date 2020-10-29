// +build go1.8

package wrapsrw

import "net/http"

var (
	_ http.Pusher = &http2FancyWriter{}
)
