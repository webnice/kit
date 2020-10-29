// +build go1.8,go1.9,go1.10,!go1.11

package wrapsrw

import "net/http"

var (
	_ http.CloseNotifier = &httpFancyWriter{}  // nolint: megacheck
	_ http.CloseNotifier = &http2FancyWriter{} // nolint: megacheck
)

func (f *httpFancyWriter) CloseNotify() <-chan bool {
	return f.basic.ResponseWriter.(http.CloseNotifier).CloseNotify() // nolint: megacheck
}

// HTTP2

func (f *http2FancyWriter) CloseNotify() <-chan bool {
	return f.basic.ResponseWriter.(http.CloseNotifier).CloseNotify() // nolint: megacheck
}
