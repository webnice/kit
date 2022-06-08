// Package test
package test

import "github.com/webnice/debug"

func (tst *impl) logPrint() {

	debug.Nop()

	tst.log().Trace("trace")
	tst.log().Debug("debug")
	tst.log().Info("info")
	tst.log().Notice("notice")
	tst.log().Warning("warning")
	tst.log().Error("error")
	tst.log().Critical("critical")
	tst.log().Alert("alert")
	tst.log().Fatality(false).Fatal("fatal")

	//tst.log().Key(kitTypes.LoggerKey{
	//	"one":   "Один.",
	//	"two":   "Два.",
	//	"three": "Три.",
	//}).Alert("||||||||||||||||||||||||||||||1234567890")

}