package trace

import (
	"strings"
	"testing"

	kitTypes "github.com/webnice/kit/v4/types"
)

func callShort(ti *kitTypes.TraceInfo, stackBack int) { Short(ti, stackBack) }

func TestShortDetailed(t *testing.T) {
	const (
		errStackTraceEmpty = "Short() не заполнил StackTrace."
		tplStackNotTrimmed = "Short() не обрезал текущий тестовый пакет:\n%s."
		tplFilenameLong    = "FilenameLong должен указывать за пределы текущего пакета: %q."
		errFilenameShort   = "FilenameShort пустой."
		tplPackageWrong    = "Package должен быть вне текущего пакета: %q."
		tplFunctionWrong   = "Function должна быть вне текущего теста: %q."
		tplLineWrong       = "Line = %d."
		currentTestName    = "TestShortDetailed"
		currentPackage     = "github.com/webnice/kit/v4/module/trace"
		currentFile        = "trace_test.go"
	)
	var ti *kitTypes.TraceInfo

	ti = kitTypes.NewTraceInfo()
	Short(ti, 0)
	if ti.StackTrace.Len() == 0 {
		t.Fatal(errStackTraceEmpty)
	}
	if strings.Contains(ti.StackTrace.String(), currentFile) {
		t.Fatalf(tplStackNotTrimmed, ti.StackTrace.String())
	}
	if ti.FilenameLong.String() == "" || strings.HasSuffix(ti.FilenameLong.String(), currentFile) {
		t.Fatalf(tplFilenameLong, ti.FilenameLong.String())
	}
	if ti.FilenameShort.String() == "" {
		t.Fatal(errFilenameShort)
	}
	if ti.Package.String() == currentPackage || ti.Package.String() == "" {
		t.Fatalf(tplPackageWrong, ti.Package.String())
	}
	if ti.Function.String() == "" || ti.Function.String() == currentTestName {
		t.Fatalf(tplFunctionWrong, ti.Function.String())
	}
	if ti.Line <= 0 {
		t.Fatalf(tplLineWrong, ti.Line)
	}
}

func TestShortStackBackDoesNotPanic(t *testing.T) {
	const (
		tplStateWrong = "Short(stackBack=1) state некорректен: line=%d filename=%q."
	)
	var ti *kitTypes.TraceInfo

	ti = kitTypes.NewTraceInfo()
	callShort(ti, 1)
	if ti.StackTrace.Len() == 0 || ti.FilenameLong.String() == "" || ti.Line < 0 {
		t.Fatalf(tplStateWrong, ti.Line, ti.FilenameLong.String())
	}
}

func TestShortNegativeStackBack(t *testing.T) {
	const (
		tplFunctionWrong = "Function с отрицательным stackBack = %q."
		tplFilenameWrong = "FilenameLong с отрицательным stackBack = %q."
		expectedFunction = "Short"
		expectedSuffix   = "trace.go"
	)
	var ti *kitTypes.TraceInfo

	ti = kitTypes.NewTraceInfo()
	Short(ti, -2)
	if ti.Function.String() != expectedFunction {
		t.Fatalf(tplFunctionWrong, ti.Function.String())
	}
	if !strings.HasSuffix(ti.FilenameLong.String(), expectedSuffix) {
		t.Fatalf(tplFilenameWrong, ti.FilenameLong.String())
	}
}

func TestFunctionCorrect(t *testing.T) {
	const (
		tplFunctionCorrectWrong = "functionCorrect(%q) = pkg:%q fn:%q."
	)
	// Описание одного теста разбора имени функции.
	type testCase struct {
		name     string
		fn       string
		pkg      string
		function string
	}
	var (
		tests []testCase
		test  testCase
		n     int
		ti    *kitTypes.TraceInfo
	)

	tests = []testCase{
		{name: "simple", fn: "simple", pkg: "", function: "simple"},
		{name: "slash", fn: "pkg/sub.Func", pkg: "pkg/sub", function: "Func"},
		{name: "method", fn: "Type.Method", pkg: "Type", function: "Method"},
		{name: "slash-method", fn: "pkg/sub.Type.Method", pkg: "pkg/sub", function: "Type.Method"},
	}
	for n = range tests {
		test = tests[n]
		t.Run(test.name, func(t *testing.T) {
			ti = kitTypes.NewTraceInfo()
			_, _ = ti.Function.WriteString(test.fn)
			functionCorrect(ti)
			if ti.Package.String() != test.pkg || ti.Function.String() != test.function {
				t.Fatalf(tplFunctionCorrectWrong, test.fn, ti.Package.String(), ti.Function.String())
			}
		})
	}
}

func TestGetStackBack(t *testing.T) {
	const (
		tplStackBackWrong = "getStackBack() = %d, ожидалось минимум 2."
	)
	var buf *buffer

	buf = getBuffer()
	defer putBuffer(buf)
	if n := getStackBack(buf); n < 2 {
		t.Fatalf(tplStackBackWrong, n)
	}
}

func TestGetStackModes(t *testing.T) {
	const (
		tplFullWrong     = "getStack(full) не вернул полный стек:\n%s."
		tplNegativeWrong = "getStack(negative stackBack) некорректен:\n%s."
		tplHugeWrong     = "getStack(huge stackBack) некорректен:\n%s."
		tplGrowthWrong   = "getStack(small buffer) некорректен:\n%s."
	)
	var (
		tiFull     *kitTypes.TraceInfo
		tiNegative *kitTypes.TraceInfo
		tiHuge     *kitTypes.TraceInfo
		tiGrowth   *kitTypes.TraceInfo
		buf        *buffer
	)

	tiFull = kitTypes.NewTraceInfo()
	buf = getBuffer()
	getStack(tiFull, 0, true, buf)
	putBuffer(buf)
	if !strings.Contains(tiFull.StackTrace.String(), "goroutine") {
		t.Fatalf(tplFullWrong, tiFull.StackTrace.String())
	}
	tiNegative = kitTypes.NewTraceInfo()
	buf = getBuffer()
	getStack(tiNegative, -2, false, buf)
	putBuffer(buf)
	if !strings.Contains(tiNegative.StackTrace.String(), "goroutine") {
		t.Fatalf(tplNegativeWrong, tiNegative.StackTrace.String())
	}
	tiHuge = kitTypes.NewTraceInfo()
	buf = getBuffer()
	getStack(tiHuge, 100000, false, buf)
	putBuffer(buf)
	if tiHuge.StackTrace.Len() == 0 || !strings.HasPrefix(tiHuge.StackTrace.String(), "goroutine") {
		t.Fatalf(tplHugeWrong, tiHuge.StackTrace.String())
	}
	tiGrowth = kitTypes.NewTraceInfo()
	buf = getBuffer()
	buf.Byte64k = make([]byte, 1)
	getStack(tiGrowth, 0, false, buf)
	putBuffer(buf)
	if tiGrowth.StackTrace.Len() == 0 || !strings.Contains(tiGrowth.StackTrace.String(), "goroutine") {
		t.Fatalf(tplGrowthWrong, tiGrowth.StackTrace.String())
	}
}

func BenchmarkShort(b *testing.B) {
	var (
		ti *kitTypes.TraceInfo
		n  int
	)

	ti = kitTypes.NewTraceInfo()
	for n = 0; n < b.N; n++ {
		Short(ti, 0)
	}
}
