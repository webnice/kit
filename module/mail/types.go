package mail

import (
	"crypto/tls"
	"net/smtp"
	"sync"
	"time"

	"github.com/webnice/kit/module/mail/encode"
	"github.com/webnice/kit/module/mail/message"
)

var smtpDialTimeout = time.Second * 10

// Interface is an interface of mail
type Interface interface {
	// NewMessage Создание нового сообщения
	NewMessage() message.Interface

	// Encoder Установка кодировщика
	Encoder(encode.Interface) Interface

	// Send Отправка сообщения
	Send(...message.Interface) error
}

// impl is an implementation of mail
type impl struct {
	smtpCfg       *SMTP            // Конфигурация доступа к SMTP серверу
	encoder       encode.Interface // Деволтовый кодировщик сообщений, передаётся в NewMessage
	isOnline      bool             // =true - соединение с SMTP сервером установлено
	smtpClient    *smtp.Client     // Соединение с SMTP сервером
	smtpTLSConfig *tls.Config      // TLS configuration
	smtpAuth      smtp.Auth        // SMTP Auth

	sync.Mutex
}

// SMTP Структура конфигурации SMTP сервера
type SMTP struct {
	Host     string `yaml:"Host"`     // Адрес почтового сервера для отправки сообщений по протоколу SMTP
	Port     uint32 `yaml:"Port"`     // Порт почтового сервера SMTP
	TLS      bool   `yaml:"TLS"`      // Протокол используемый для подключения к почтовому серверу. =true-TLS, =false-без шифрования
	Source   string `yaml:"Source"`   // Адрес электронной почты с которого сервер отправляет все сообщения по умолчанию
	Username string `yaml:"Username"` // Имя пользователя - Реквизиты доступа к серверу
	Password string `yaml:"Password"` // Пароль - Реквизиты доступа к серверу
	Template string `yaml:"Template"` // Папка шаблонов писем
}
