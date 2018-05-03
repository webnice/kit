package verify

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import "gopkg.in/webnice/web.v1/context"
import ()

func init() {
	// Регистрация внешней функции проверки данных
	context.New().RegisterVerifyPlugin(new(webContextVerify))
}

type webContextVerify struct{}

// Verify Implementation of context.VerifyPlugin
func (wcv *webContextVerify) Verify(data interface{}) (rsp []byte, err error) {
	var vfi Interface

	if vfi, err = Verify(data); err != nil {
		rsp = vfi.Message(err.Error()).Json()
		return
	}

	return
}
