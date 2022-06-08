// Package logger
package log

import (
	"time"

	kitModuleLogLevel "github.com/webnice/kit/module/log/level"
	kitModuleTrace "github.com/webnice/kit/module/trace"
)

// Write Реализация интерфейса io.Writer.
func (log *logger) Write(b []byte) (n int, err error) {
	const stackBack = 3
	var msg *Message

	msg = log.MessageGet()
	msg.Timestamp = time.Now().In(time.UTC)
	msg.Level = findLoglevel()
	msg.Fatality = msg.Level == kitModuleLogLevel.Fatal
	kitModuleTrace.Short(msg.Trace, stackBack)
	n, err = msg.Pattern.Write(rexSpaceLast.ReplaceAll(rexSpaceFirst.ReplaceAll(b, []byte{}), []byte{}))
	log.Message(msg)

	return
}

// WriteString Реализация интерфейса io.StringWriter.
func (log *logger) WriteString(s string) (n int, err error) {
	const stackBack = 3
	var msg *Message

	msg = log.MessageGet()
	msg.Timestamp = time.Now().In(time.UTC)
	msg.Level = findLoglevel()
	msg.Fatality = msg.Level == kitModuleLogLevel.Fatal
	kitModuleTrace.Short(msg.Trace, stackBack)
	n, err = msg.Pattern.WriteString(rexSpaceLast.ReplaceAllString(rexSpaceFirst.ReplaceAllString(s, ``), ``))
	log.Message(msg)

	return
}
