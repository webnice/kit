// +build go1.11

package wrapsrw

import (
	"io"
	"net/http"
)

// New Proxy around an http.ResponseWriter that allows you to hook into various parts of the response process
func New(wr http.ResponseWriter, protoMajor int) WrapsResponseWriter {
	var fl, ps, hj, rf bool
	var ba = basic{
		ResponseWriter: wr,
	}

	_, fl = wr.(http.Flusher)
	switch protoMajor {
	case 2:
		_, ps = wr.(http.Pusher)
		if fl && ps {
			return &http2FancyWriter{ba}
		}
	default:
		_, hj = wr.(http.Hijacker)
		_, rf = wr.(io.ReaderFrom)
		if fl && hj && rf {
			return &httpFancyWriter{ba}
		}
	}
	if fl {
		return &flush{ba}
	}

	return &ba
}
