package auth // import "github.com/webnice/kit/v1/modules/mail/auth"

import (
	"bytes"
	"fmt"
	"net/smtp"
)

var (
	errUnencryptedConnection     = fmt.Errorf(`unencrypted connection`)
	errWrongServerHostName       = fmt.Errorf(`wrong server host name`)
	errUnexpectedServerChallenge = fmt.Errorf(`unexpected server challenge`)
)

// Interface auth interface
type Interface interface {
	UserName(string) Interface                      // Set username
	Password(string) Interface                      // Set password
	HostName(string) Interface                      // Set hostname
	Start(*smtp.ServerInfo) (string, []byte, error) // Start authorization
	Next([]byte, bool) ([]byte, error)              // Next authorization steps
}

// impl auth implementation
type impl struct {
	userName string
	password string
	hostName string
}

// ErrUnencryptedConnection Unencrypted connection error
func ErrUnencryptedConnection() error { return errUnencryptedConnection }

// ErrWrongServerHostName Wrong server host name error
func ErrWrongServerHostName() error { return errWrongServerHostName }

// ErrUnexpectedServerChallenge Unexpected server challenge error
func ErrUnexpectedServerChallenge(s string) error {
	return fmt.Errorf(errUnexpectedServerChallenge.Error()+": %s", s)
}

// New interface
func New() Interface {
	var au = new(impl)
	return au
}

// UserName set username
func (au *impl) UserName(userName string) Interface {
	au.userName = userName
	return au
}

// Password set password
func (au *impl) Password(password string) Interface {
	au.password = password
	return au
}

// HostName set hostname
func (au *impl) HostName(hostName string) Interface {
	au.hostName = hostName
	return au
}

// Start authorization
func (au *impl) Start(srv *smtp.ServerInfo) (ret string, data []byte, err error) {
	var mechanism string

	if !srv.TLS {
		for _, mechanism = range srv.Auth {
			if mechanism == "LOGIN" {
				ret = mechanism
				break
			}
		}
		if ret == "" {
			err = ErrUnencryptedConnection()
			return
		}
	}
	if srv.Name != au.hostName {
		err = ErrWrongServerHostName()
		return
	}
	return
}

// Next authorization steps
func (au *impl) Next(from []byte, more bool) (data []byte, err error) {
	if !more {
		return
	}
	switch {
	case bytes.Equal(from, []byte("Username:")):
		data = []byte(au.userName)
		return
	case bytes.Equal(from, []byte("Password:")):
		data = []byte(au.password)
		return
	default:
		err = ErrUnexpectedServerChallenge(string(from))
		return
	}
	return
}
