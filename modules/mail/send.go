package mail // import "gopkg.in/webnice/kit.v1/modules/mail"

//import "gopkg.in/webnice/debug.v1"
import "gopkg.in/webnice/log.v2"
import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strings"

	"gopkg.in/webnice/kit.v1/modules/mail/auth"
	"gopkg.in/webnice/kit.v1/modules/mail/message"
)

// Send Отправка сообщения
func (ml *impl) Send(messages ...message.Interface) (err error) {
	var msg message.Interface
	var i int

	ml.Lock()
	defer ml.Unlock()

	if !ml.isOnline {
		if ml.smtpClient, err = ml.makeSMTPClient(); err != nil {
			return
		}
		ml.isOnline = true
	} else {
		// Сбрасываем состояние предыдущей работы с сервером
		_ = ml.smtpClient.Reset()
	}

	// Отправка всех сообщений через одно соединение
	for i, msg = range messages {
		var wr io.WriteCloser

		// Если сообщений больше одного
		if i > 0 {
			log.Notice("mail client Reset()")
			if err = ml.smtpClient.Reset(); err != nil {
				return
			}
		}

		//log.Noticef("From: '%s'", msg.GetFrom())
		if err = ml.smtpClient.Mail(msg.GetFrom()); err != nil {
			return
		}
		//log.Noticef("To: '%s'", msg.GetTo())
		if err = ml.smtpClient.Rcpt(msg.GetTo()); err != nil {
			return
		}
		if wr, err = ml.smtpClient.Data(); err != nil {
			return
		}
		if _, err = msg.WriteTo(wr); err != nil {
			return
		}
		err = wr.Close()
	}

	return
}

// tls TLS configuration
func (ml *impl) tls() *tls.Config {
	if ml.smtpTLSConfig == nil {
		ml.smtpTLSConfig = &tls.Config{
			ServerName: ml.smtpConfiguration.SMTPServer,
		}
	}
	return ml.smtpTLSConfig
}

// makeSMTPClient Содание SMTP клиента
func (ml *impl) makeSMTPClient() (client *smtp.Client, err error) {
	var conn net.Conn
	var ok bool

	if conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ml.smtpConfiguration.SMTPServer, ml.smtpConfiguration.SMTPPort), smtpDialTimeout); err != nil {
		return
	}
	if ml.smtpConfiguration.SMTPTLS {
		conn = tls.Client(conn, ml.tls())
	}
	if client, err = smtp.NewClient(conn, ml.smtpConfiguration.SMTPServer); err != nil {
		return
	}
	// HELLO SMTP command
	if err = client.Hello("localhost"); err != nil {
		return
	}
	if !ml.smtpConfiguration.SMTPTLS {
		if ok, _ = client.Extension(`STARTTLS`); ok {
			if err = client.StartTLS(ml.tls()); err != nil {
				_ = client.Close()
				return
			}
		}
	}

	// Функция авторизации
	if ml.smtpAuth == nil && ml.smtpConfiguration.AuthUserName != "" {
		var auths string
		if ok, auths = client.Extension("AUTH"); ok {
			if strings.Contains(auths, "CRAM-MD5") {
				ml.smtpAuth = smtp.CRAMMD5Auth(ml.smtpConfiguration.AuthUserName, ml.smtpConfiguration.AuthPassword)
			} else if strings.Contains(auths, "LOGIN") && !strings.Contains(auths, "PLAIN") {
				ml.smtpAuth = auth.New().
					UserName(ml.smtpConfiguration.AuthUserName).
					Password(ml.smtpConfiguration.AuthPassword).
					HostName(ml.smtpConfiguration.SMTPServer)
			} else {
				ml.smtpAuth = smtp.PlainAuth("", ml.smtpConfiguration.AuthUserName, ml.smtpConfiguration.AuthPassword, ml.smtpConfiguration.SMTPServer)
			}
		}
	}

	// Авторизация на SMTP сервере
	if ml.smtpAuth != nil {
		if err = client.Auth(ml.smtpAuth); err != nil {
			_ = client.Close()
			return
		}

	}

	return
}
