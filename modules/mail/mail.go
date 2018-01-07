package mail // import "gopkg.in/webnice/kit.v1/modules/mail"

//import "gopkg.in/webnice/debug.v1"
import "gopkg.in/webnice/log.v2"
import (
	"runtime"
	"strings"

	"gopkg.in/webnice/kit.v1/modules/mail/encode"
	"gopkg.in/webnice/kit.v1/modules/mail/message"
)

// New return mail interface
func New(cfg *SmtpConfiguration) Interface {
	var ml = new(impl)
	ml.smtpConfiguration = cfg
	ml.encoder = encode.New()
	runtime.SetFinalizer(ml, destructor)
	return ml
}

// checkOKError Проверка на "250 2.0.0 Ok"
func checkOKError(inp error) (err error) {
	if inp == nil {
		return
	}
	if strings.Index(strings.TrimSpace(inp.Error()), "250") != 0 {
		err = inp
	}
	return
}

// destructor Object destructor
func destructor(ml *impl) {
	var err error
	if ml.isOnline && ml.smtpClient != nil {
		if err = checkOKError(ml.smtpClient.Quit()); err != nil {
			log.Warningf("Error quit smtp session: %s", err.Error())
		}
		_ = ml.smtpClient.Close()
	}
}

// NewMessage Создание нового сообщения
func (ml *impl) NewMessage() message.Interface {
	var msg = message.New()
	msg.Encoder(ml.encoder)
	if ml.smtpConfiguration != nil {
		msg.From(ml.smtpConfiguration.FromAddress)
	}
	return msg
}

// Encoder Set encoder
func (ml *impl) Encoder(encoder encode.Interface) Interface {
	ml.encoder = encoder
	return ml
}
