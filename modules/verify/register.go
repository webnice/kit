package verify

import "github.com/webnice/web/v2/context"

func init() {
	// Регистрация внешней функции проверки данных
	context.RegisterGlobalVerifyPlugin(new(webContextVerify))
}

type webContextVerify struct{}

// Verify Implementation of context.VerifyPlugin
func (wcv *webContextVerify) Verify(data interface{}) (rsp []byte, err error) {
	var vfi Interface

	if vfi, err = Verify(data); err != nil {
		if vfi != nil {
			rsp = vfi.Message(err.Error()).Json()
		}
		return
	}

	return
}

// Verify Implementation of context.VerifyPlugin
func (wcv *webContextVerify) Error400(err error) (rsp []byte) {
	const defaultError = `unknown error`

	if err != nil {
		rsp = E4xx().Code(-1).Message(err.Error()).Json()
	} else {
		rsp = E4xx().Code(-1).Message(defaultError).Json()
	}

	return
}
