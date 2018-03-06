package ch // import "gopkg.in/webnice/kit.v1/modules/ch"

//import "gopkg.in/webnice/debug.v1"
//import "gopkg.in/webnice/log.v2"
import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// MakeDsn Создание строки подключения к базе данных для database/sql подобных драйверов
func (db *Implementation) MakeDsn() (err error) {
	var cnf *Configuration = db.cnf
	var u *url.URL
	var v url.Values

	// Защита от паники
	if cnf == nil {
		cnf = defaultConfiguration
	}
	if cnf == nil {
		err = fmt.Errorf("Configuration is empty")
		return
	}
	// Хосты
	if len(cnf.Hosts) <= 0 {
		err = fmt.Errorf("Database hosts are not specified")
		return
	}
	// Создание DSN на базе URL
	u, _ = url.Parse(`tcp://` + cnf.Hosts[0])
	v = u.Query()
	if len(cnf.Hosts) > 1 {
		v.Add(`alt_hosts`, strings.Join(cnf.Hosts[1:], `,`))
	}
	// Имя пользователя и пароль
	v.Add(`username`, cnf.Login)
	if cnf.Password != "" {
		v.Add(`password`, cnf.Password)
	}
	// База данных по умолчанию
	if cnf.Database != "" {
		v.Add(`database`, cnf.Database)
	}
	v.Add(`read_timeout`, fmt.Sprintf("%d", cnf.ReadTimeout/time.Second))
	v.Add(`write_timeout`, fmt.Sprintf("%d", cnf.WriteTimeout/time.Second))
	v.Add(`no_delay`, fmt.Sprintf("%t", cnf.NoDelay))
	v.Add(`compress`, fmt.Sprintf("%t", cnf.Compress))
	v.Add(`connection_open_strategy`, cnf.OpenStrategy)
	if cnf.BlockSize > 0 {
		v.Add(`block_size`, fmt.Sprintf("%d", cnf.BlockSize))
	}
	v.Add(`secure`, fmt.Sprintf("%t", cnf.Secure))
	v.Add(`skip_verify`, fmt.Sprintf("%t", cnf.SkipVerify))
	if cnf.Debug {
		v.Add(`debug`, `true`)
	}
	u.RawQuery = v.Encode()
	// Получение готовой строки
	db.dsn = u.String()

	return
}

// Dsn Return DSN string
func (db *Implementation) Dsn() (ret string, err error) {
	if db.dsn != "" {
		ret = db.dsn
		return
	}
	if err = db.MakeDsn(); err != nil {
		return
	}
	ret = db.dsn
	return
}
