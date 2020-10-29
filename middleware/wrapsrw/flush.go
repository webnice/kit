package wrapsrw

import "net/http"

func (f *flush) Flush() { f.basic.ResponseWriter.(http.Flusher).Flush() }
