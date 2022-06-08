// Package migrations
package migrations

//import (
//	//"bytes"
//	//"fmt"
//	//"os/exec"
//	//
//	//"application/modules/launcher"
//	//
//	//"github.com/webnice/kit/modules/ch"
//)

// Поиск и проверка версии утилиты миграции.
func (migrations *impl) migrationsUtility() (ret string) {
	//const command = `gsmigrate`
	//var err error
	//
	//if ret, err = exec.LookPath(command); err != nil {
	//	log.Warningf("can't find migrations utility: %s", err.Error())
	//	return
	//}

	return
}

func (migrations *impl) migrationsSQL(command string) (err error) {
	//var dsn string
	//
	//// Если не указана директория с миграциями, то выход
	//if migrations.Cfg.Configuration().Database.Migrations == "" {
	//	log.Warningf("folder with mysql database migration files is not specified. Migrations are not not applied")
	//	return
	//}
	//if dsn, err = db.New().Dsn(); err != nil {
	//	err = fmt.Errorf("database configuration error: %s", err.Error())
	//	return
	//}
	//// Применение миграций
	//if err = migrations.migrationsApply(
	//	command,
	//	migrations.Cfg.Configuration().Database.Migrations,
	//	migrations.Cfg.Configuration().Database.Driver,
	//	dsn,
	//); err != nil {
	//	log.Warningf("migrations warnings: %s", err.Error())
	//	return
	//}

	return
}

//func (mrs *impl) migrationsClickhouse(command string) (err error) {
//	var dsn string
//
//	// Если не указана директория с миграциями, то выход
//	if mrs.Cfg.Configuration().Clickhouse.Migrations == "" {
//		log.Warningf("folder with clickhouse database migration files is not specified. Migrations are not not applied")
//		return
//	}
//	if dsn, err = ch.New().Dsn(); err != nil {
//		err = fmt.Errorf("database configuration error: %s", err.Error())
//		return
//	}
//	// Применение миграций
//	if err = mrs.migrationsApply(command, mrs.Cfg.Configuration().Clickhouse.Migrations, "clickhouse", dsn); err != nil {
//		log.Warningf("migrations warnings: %s", err.Error())
//		return
//	}
//
//	return
//}

// Применение миграций.
func (migrations *impl) migrationsApply(command string, dir string, drv string, dsn string) (err error) {
	//var lau launcher.Interface
	//var cmd, env []string
	//var ecode int
	//var oBuf, eBuf *bytes.Buffer
	//
	//env = []string{`GOOSE_DIR=` + dir, `GOOSE_DRV=` + drv, `GOOSE_DSN=` + dsn}
	//cmd = []string{command, `up`}
	//oBuf, eBuf = &bytes.Buffer{}, &bytes.Buffer{}
	//lau = launcher.New()
	//if err = lau.Launch(cmd, env, nil, oBuf, eBuf); err != nil {
	//	return
	//}
	//if ecode, err = lau.Wait(); err != nil || ecode != 0 {
	//	err = fmt.Errorf("utility %q exit with error code %d: %s", command, ecode, err.Error())
	//}
	//if oBuf.Len() > 0 {
	//	log.Noticef("migration utility (out): %q", oBuf.String())
	//}
	//if eBuf.Len() > 0 {
	//	log.Warningf("migration utility (err): %q", eBuf.String())
	//}

	return
}
