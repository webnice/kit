package response

import (
//	"net/http"
//	"reflect"
//
//	"gopkg.in/webnice/web.v1/header"
//	"gopkg.in/webnice/web.v1/mime"
)

//func indirect(rv reflect.Value) reflect.Value {
//	for rv.Kind() == reflect.Ptr {
//		rv = rv.Elem()
//	}
//	return rv
//}
//
//// Function returns empty array if v is a nil slice
//func normalizeArrayIfNeeded(v interface{}) interface{} {
//	var val = indirect(reflect.ValueOf(v))
//	if (val.Kind() == reflect.Array || val.Kind() == reflect.Slice) && val.Len() == 0 {
//		return make([]int, 0)
//	}
//	return v
//}
//
//// JSON Encoding an object to json format and printing the result with header and status code
//func JSON(wr http.ResponseWriter, st int, val interface{}) (err error) {
//	wr.Header().Add(header.ContentType, mime.ApplicationJSONCharsetUTF8)
//	wr.WriteHeader(st)
//	err = json.NewEncoder(wr).Encode(val)
//	return
//}
