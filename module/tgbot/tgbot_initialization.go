package tgbot

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/webnice/tba/v9"
)

// LogOut Выполнение процедуры завершения работы телеграм бота и обработки флага выхода с сервера.
func (tbt *impl) LogOut() (err error) {
	const (
		errRemove  = "Удаление файл-флага прервано ошибкой: %s."
		logCommand = "Выполнение команды \"LogOut\" для телеграм бота."
		logOk      = "Команда \"LogOut\" успешно выполнена."
	)
	var (
		lut tgbotapi.Chattable
		rsp *tgbotapi.APIResponse
		fi  os.FileInfo
	)

	//

	//

	var pcb tgbotapi.SavePreparedKeyboardButtonConfig
	pcb = tgbotapi.NewSavePreparedKeyboardButton(1, tgbotapi.NewKeyboardButton("Текст"))
	spkb, _ := tbt.api.SavePreparedKeyboardButton(pcb)

	_ = spkb

	//

	//

	if err = tbt.initializationDeleteWebhook(); err != nil {
		tbt.log().Warning(err)
	}
	// Выполнение алгоритма обработки файл-флага.
	if fi, err = os.Stat(tbt.botCfg.LogoutFlagFile); err != nil {
		err = nil
		return
	}
	if fi.IsDir() {
		return
	}
	if err = os.Remove(tbt.botCfg.LogoutFlagFile); err != nil {
		tbt.log().Warningf(errRemove, err)
		err = nil
		return
	}
	// Файл-флаг был и успешно удалён.
	// Выполнение команды LogOut.
	lut = new(tgbotapi.LogOutConfig)
	tbt.log().Notice(logCommand)
	if rsp, err = tbt.api.Request(lut); err != nil {
		return
	}
	if rsp.Ok {
		tbt.log().Notice(logOk)
	}

	return
}

// Initialization Инициализация и запуск телеграм бота.
func (tbt *impl) Initialization(ctx context.Context, cfg *Configuration) (err error) {
	const (
		msgInfo      = "Телеграм бот авторизован под псевдонимом %q."
		errLoggedOut = "Logged out"
		timeoutTic1  = time.Minute * 1
	)
	var (
		onDone  chan struct{}
		hClient *http.Client
		end     bool
		first   bool
		tic1    *time.Ticker
	)

	// Связь через модифицированный http клиент.
	hClient = tbt.initializationCfgHttpClient(cfg)
	tic1, first = time.NewTicker(time.Second/4), true
	defer tic1.Stop()
	for !end {
		select {
		case <-ctx.Done():
			end = true
			continue
		case <-tic1.C:
		}
		switch hClient != nil {
		case true: // Создание клиента телеграм с модифицированным транспортом.
			tbt.api, err = tgbotapi.NewBotAPIWithClient(cfg.Token, cfg.UriApi, cfg.UriFile, hClient)
		default: // Создание клиента телеграм со стандартным транспортом.
			tbt.api, err = tgbotapi.NewBotAPIWithAPIEndpoint(cfg.Token, cfg.UriApi, cfg.UriFile)
		}
		if first { // Изменение таймера.
			tic1.Stop()
			tic1, first = time.NewTicker(timeoutTic1), false
		}
		if err != nil && strings.EqualFold(err.Error(), errLoggedOut) {
			tbt.log().Notice("Ожидание до 10 минут.")
			continue
		}
		end = true
	}
	if err != nil {
		return
	}
	tbt.botCfg, tbt.botUser = cfg, &tbt.api.Self
	tbt.log().Infof(msgInfo, tbt.api.Self.UserName)
	// Подготовка способа получения обновлений от сервера телеграм.
	switch tbt.botCfg.Webhook {
	case "":
		_ = tbt.initializationDeleteWebhook()
		onDone = make(chan struct{})
		go tbt.goRequestUpdate(onDone, ctx)
		<-onDone
		close(onDone)
	case "-":
		// Данные от серверов телеграм не запрашиваются и не принимаются.
		// Бот может только отправлять сообщения.
	default:
		tbt.initializationRegistrationWebhook()
	}
	onDone = make(chan struct{})
	// Запуск потока загрузки обновлений.
	go tbt.goNewTelegramMessage(onDone, ctx)
	<-onDone
	close(onDone)
	tbt.botReady = true

	return
}

// Инициализация клиента HTTP соединения для подключения к серверу телеграм.
func (tbt *impl) initializationCfgHttpClient(cfg *Configuration) (ret *http.Client) {
	const (
		pTcp4, pTcp6 = "tcp4", "tcp6"
		infProxy     = "настроена связь с telegram.org с использованием прокси: %s"
		infNetwork   = "настроена связь с telegram.org с использованием сети: %s"
	)
	var (
		transport  *http.Transport
		zeroDialer net.Dialer
		proxy      *url.URL
	)

	// Транспорт в зависимости от настроенного семейства протоколов.
	switch proxy = tbt.initializationCfgProxy(cfg.Proxy); cfg.NetworkFamily {
	case pTcp4, pTcp6:
		// Подключаться только по указанному семейству протоколов.
		transport = http.DefaultTransport.(*http.Transport).Clone()
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return zeroDialer.DialContext(ctx, cfg.NetworkFamily, addr)
		}
		// Добавление прокси, если указан.
		if proxy != nil {
			transport.Proxy = http.ProxyURL(proxy)
			tbt.log().Infof(infProxy, cfg.Proxy)
		}
		// Объект http клиента.
		ret = &http.Client{Transport: transport}
		tbt.log().Infof(infNetwork, cfg.NetworkFamily)
	}
	// Транспорт с использованием только прокси.
	if ret == nil && proxy != nil {
		ret = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
		tbt.log().Infof(infProxy, cfg.Proxy)
	}

	return
}

// Инициализация адреса прокси сервера.
func (tbt *impl) initializationCfgProxy(proxy string) (ret *url.URL) {
	var err error

	if proxy == "" {
		return
	}
	if ret, err = url.Parse(proxy); err != nil {
		err = tbt.Errors().ProxyError.Bind(proxy, err)
		tbt.log().Critical(err)
	}

	return
}

// Отмена настройки телеграм сервера, заставляющего телеграм сервер присылать вебхуки с сообщениями для телеграм бота.
func (tbt *impl) initializationDeleteWebhook() (err error) {
	const infDeleteWebhook = "Удаление вебхука %q телеграм бота %q."
	var webhookInfo tgbotapi.WebhookInfo

	if webhookInfo, err = tbt.api.GetWebhookInfo(); err != nil {
		tbt.log().Error(tbt.Errors().GetWebhookInfo.Bind(err))
		return
	}
	if !webhookInfo.IsSet() {
		return
	}
	// Отправка запроса удаления регистрации вебхука.
	tbt.log().Infof(infDeleteWebhook, webhookInfo.URL, tbt.botUser.UserName)
	_, err = tbt.api.Request(tgbotapi.DeleteWebhookConfig{})

	return
}

// Создание URI адреса получения вебхук от сервера телеграм.
func (tbt *impl) initializationTokenSha256() (ret *url.URL, err error) {
	var sum hash.Hash

	sum = sha256.New()
	sum.Write([]byte(tbt.botCfg.Token))
	if ret, err = url.
		Parse(fmt.Sprintf("%s/%x", tbt.botCfg.Webhook, sum.Sum(nil))); err != nil {
		return
	}

	return
}

// Регистрация телеграм бота в режиме получения сообщений через вебхук от сервера телеграм.
func (tbt *impl) initializationRegistrationWebhook() {
	var (
		err error
		whc tgbotapi.WebhookConfig
		whs *tgbotapi.APIResponse
		uri *url.URL
	)

	if uri, err = tbt.
		initializationTokenSha256(); err != nil {
		tbt.log().Fatal(tbt.Errors().WebhookCreate.Bind(err))
		return
	}
	if whc, err = tgbotapi.
		NewWebhook(uri.String()); err != nil {
		tbt.log().Error(tbt.Errors().WebhookRegistration.Bind(uri.String(), err))
	}
	// Подписка на сообщения которые будут приходить через вызов вебхука.
	whc.AllowedUpdates = tbt.defaultAllowedUpdates()
	if err = requestRetryAfter(func() (ok bool, st int64, ra int64, er error) {
		if whs, er = tbt.api.Request(whc); whs == nil {
			return
		}
		if ok, st, er = whs.Ok, whs.ErrorCode, nil; whs.Parameters != nil {
			ra = whs.Parameters.RetryAfter
		}
		return
	}); err != nil {
		tbt.log().Fatal(err.Error())
		return
	}
}

// Горутина телеграм бота в режиме получения сообщений через периодический опрос api телеграм сервера.
func (tbt *impl) goRequestUpdate(onDone chan<- struct{}, ctx context.Context) {
	var (
		end bool
		upd tgbotapi.UpdateConfig
		uch tgbotapi.UpdatesChannel
		msg *tgbotapi.Update
	)

	upd = tgbotapi.NewUpdate(0)
	upd.AllowedUpdates, upd.Timeout = tbt.defaultAllowedUpdates(), getUpdatesRequestTimeout
	uch = tbt.api.GetUpdatesChan(upd)
	onDone <- struct{}{}
	for !end {
		select {
		case <-ctx.Done(): // Канал получения прерывания.
			end = true
			continue
		case update := <-uch: // Канал получения обновлений с сервера телеграм.
			msg = &update
			// Отправка полученного обновления в шину данных.
			_, _ = tbt.Consumer(false, msg)
		}
	}
}

// Горутина загрузки обновлений от телеграм сервера.
// Обновления загружаются из шины данных, в которую они попадают двумя путями:
//  1. Из опроса сервера телеграм, при работе через RequestUpdate.
//  2. Из вебхук, при работе через вебхук.
func (tbt *impl) goNewTelegramMessage(onDone chan<- struct{}, ctx context.Context) {
	var (
		end bool
		upd *tgbotapi.Update
	)

	onDone <- struct{}{}
	for !(end && len(tbt.msgInp) == 0) {
		select {
		case <-ctx.Done():
			end = true
			continue
		case upd = <-tbt.msgInp:
			// Единая точка входа всех новых обновлений от сервера телеграм, полученных разными способами.
			// Далее, для каждого обновления создаётся отдельный поток, для осуществления параллельной обработки.
			go serverHandler{tbt}.ServeTelegram(upd)
		}
	}
}
