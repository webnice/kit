package mail

/*
Пример использования

import "github.com/webnice/kit/modules/mail"
import "log"
import "io/ioutil"
import "bytes"

// Конфигурация подключения к SMTP серверу
var cfg = &mail.SmtpConfiguration{
	FromAddress:  "my-address@example.com",
	SMTPServer:   "smtp.yandex.ru",
	SMTPPort:     465,
	SMTPTLS:      true,
	AuthUserName: "my-address@example.com",
	AuthPassword: "myPassword",
}

func main() {
	sender := mail.New(cfg)
	sender.Send(
		sender.NewMessage().To("toaddress@example.com").Body("Hello world").Subject("My subject"),
		sender.NewMessage().To("toaddress@example.com").ToCC("second@example.com").Body("Second message").Subject("My subject").Header("X-Custom-Name", "Custom header"),
	)

	//...

	sender := mail.New(&cnf)
	message := sender.NewMessage().Subject("My subject")
	message.To(message.EmailAddress("address@example.com", "Name Lastname Middlename"))
	message.Header("MyCustomHeaderName", "Custom header body")

	// Встраиваемые в html файлы:
	buf, err := ioutil.ReadFile("filename.png")
	message.Embed(bytes.NewReader(buf), "EmbededFileName.png", "image/png").Body(`<html>
<body>
Встраиваемый файл: '<img src="cid:EmbededFileName.png">'<br>
</body>
</html>`)

	// Все файлы (io.Reader) считываются только в момент вызова sender.Send или message.WriteTo
	if err := sender.Send(message); err != nil {
		log.Fatalf("Error send e-mail: %s", err.Error())
	}


	// В момент уничтожения объекта происходит разрыв соединения с SMTP сервером
	sender=nil
}

*/
