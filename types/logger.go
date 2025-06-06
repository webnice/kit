package types

import (
	"time"

	kitModuleLogLevel "github.com/webnice/kit/v4/module/log/level"
)

// Logger Интерфейс доступа к методам логирования.
type Logger interface {
	// Функции уточнения параметров.

	// Key Ключи логирования дополняющие лог.
	Key(keys ...LoggerKey) Logger

	// Time Переопределение времени записи переданным значением.
	Time(timestamp time.Time) Logger

	// StackBackCorrect Дополнительная коррекция просмотра стека вызовов для поиска функции вызвавшей логирование.
	// Используется когда целевая функция в стеке вызова смещена больше или меньше стандартного значения.
	StackBackCorrect(stepBack int) Logger

	// Fatality Изменение режима фатальности по умолчанию для записи лога.
	// Допустимые значения:
	// * true  - после вывода записи в лог, приложение получает сигнал немедленного завершения.
	// * false - отменяет завершение приложения, указанное для уровня логирования по умолчанию, например,
	//           для записи лога Fatal, можно отменить завершение приложения: log.Fatality(false).Fatal(...).
	Fatality(fy bool) Logger

	// Универсальна функция логирования.

	// MessageWithLevel Отправка сообщения в лог с указанием уровня логирования.
	MessageWithLevel(level kitModuleLogLevel.Level, pattern string, args ...any)

	// Конечные функции

	// Fatal Уровень 0: уровень предсмертных сообщений - система не стабильна, продолжение работы невозможно.
	Fatal(args ...any)

	// Fatalf Уровень 0: уровень предсмертных сообщений - система не стабильна, продолжение работы невозможно.
	Fatalf(pattern string, args ...any)

	// Alert Уровень 1: уровень сообщений тревоги - система нестабильна, но может частично продолжить работу.
	Alert(args ...any)

	// Alertf Уровень 1: уровень сообщений тревоги - система нестабильна, но может частично продолжить работу.
	Alertf(pattern string, args ...any)

	// Critical Уровень 2: уровень критических ошибок - часть функционала системы работает не корректно.
	Critical(args ...any)

	// Criticalf Уровень 2: уровень критических ошибок - часть функционала системы работает не корректно.
	Criticalf(pattern string, args ...any)

	// Error Уровень 3: уровень не критических ошибок - ошибки не прерывающие работу приложения.
	Error(args ...any)

	// Errorf Уровень 3: уровень не критических ошибок - ошибки не прерывающие работу приложения.
	Errorf(pattern string, args ...any)

	// Warning Уровень 4: уровень сообщений с предупреждениями.
	Warning(args ...any)

	// Warningf Уровень 4: уровень сообщений с предупреждениями.
	Warningf(pattern string, args ...any)

	// Notice Уровень 5: уровень штатных информационных сообщений, требующих повышенного внимания.
	Notice(args ...any)

	// Noticef Уровень 5: уровень штатных информационных сообщений, требующих повышенного внимания.
	Noticef(pattern string, args ...any)

	// Info Уровень 6: сообщения информационного характера описывающие шаги выполнения алгоритмов приложения.
	Info(args ...any)

	// Infof Уровень 6: сообщения информационного характера описывающие шаги выполнения алгоритмов приложения.
	Infof(pattern string, args ...any)

	// Debug Уровень 7: уровень отладочных сообщений.
	Debug(args ...any)

	// Debugf Уровень 7: уровень отладочных сообщений.
	Debugf(pattern string, args ...any)

	// Trace Уровень 8: уровень максимально подробной трассировки.
	Trace(args ...any)

	// Tracef Уровень 8: уровень максимально подробной трассировки.
	Tracef(pattern string, args ...any)
}

// LoggerKey Ключи логирования.
type LoggerKey map[string]any
