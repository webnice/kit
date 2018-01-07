package mail // import "gopkg.in/webnice/kit.v1/modules/mail"

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import (
	"crypto/tls"
	"net/smtp"
	"sync"
	"time"

	"gopkg.in/webnice/kit.v1/modules/mail/encode"
	"gopkg.in/webnice/kit.v1/modules/mail/message"
)

var smtpDialTimeout = time.Second * 10

// Interface is an interface of mail
type Interface interface {
	NewMessage() message.Interface      // Создание нового сообщения
	Encoder(encode.Interface) Interface // Set encoder
	Send(...message.Interface) error    // Отправка сообщения
}

// impl is an implementation of mail
type impl struct {
	smtpConfiguration *SmtpConfiguration // Конфигурация доступа к SMTP серверу
	encoder           encode.Interface   // Деволтовый кодировщик сообщений, передаётся в NewMessage
	isOnline          bool               // =true - соединение с SMTP сервером установлено
	smtpClient        *smtp.Client       // Соединение с SMTP сервером
	smtpTLSConfig     *tls.Config        // TLS configuration
	smtpAuth          smtp.Auth          // SMTP Auth

	sync.Mutex
}

// SmtpConfiguration Структура описания конфигурации SMTP сервера для отправки почты
type SmtpConfiguration struct {
	FromAddress  string `yaml:"FromAddress"`  // Адрес электронной почты с которого сервер отправляет все сообщения по умолчанию
	SMTPServer   string `yaml:"SMTPServer"`   // Адрес почтового сервера для отправки сообщений по протоколу SMTP
	SMTPPort     uint32 `yaml:"SMTPPort"`     // Порт почтового сервера SMTP
	SMTPTLS      bool   `yaml:"SMTPTLS"`      // Протокол используемый для подключения к почтовому серверу. =true-TLS, =false-без шифрования
	AuthUserName string `yaml:"AuthUserName"` // Имя пользователя - Реквизиты доступа к серверу
	AuthPassword string `yaml:"AuthPassword"` // Пароль - Реквизиты доступа к серверу
}
