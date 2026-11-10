package trace

import (
	"strings"
	"testing"

	kitTypes "github.com/webnice/kit/v4/types"
)

func callShort(ti *kitTypes.TraceInfo, stackBack int) { Short(ti, stackBack) }

func TestShortDetailed(t *testing.T) {
	ti := kitTypes.NewTraceInfo()
	Short(ti, 0)
	if ti.StackTrace.Len() == 0 {
		t.Fatal("Short() не заполнил StackTrace")
	}
	if strings.Contains(ti.StackTrace.String(), "trace_test.go") {
		t.Fatalf("Short() не обрезал текущий тестовый пакет:\n%s", ti.StackTrace.String())
	}
	if ti.FilenameLong.String() == "" || strings.HasSuffix(ti.FilenameLong.String(), "trace_test.go") {
		t.Fatalf("FilenameLong должен указывать за пределы текущего пакета: %q", ti.FilenameLong.String())
	}
	if ti.FilenameShort.String() == "" {
		t.Fatal("FilenameShort пустой")
	}
	if ti.Package.String() == "github.com/webnice/kit/v4/module/trace" || ti.Package.String() == "" {
		t.Fatalf("Package должен быть вне текущего пакета: %q", ti.Package.String())
	}
	if ti.Function.String() == "" || ti.Function.String() == "TestShortDetailed" {
		t.Fatalf("Function должна быть вне текущего теста: %q", ti.Function.String())
	}
	if ti.Line <= 0 {
		t.Fatalf("Line = %d", ti.Line)
	}
}

func TestShortStackBackDoesNotPanic(t *testing.T) {
	ti := kitTypes.NewTraceInfo()
	callShort(ti, 1)
	if ti.StackTrace.Len() == 0 || ti.FilenameLong.String() == "" || ti.Line < 0 {
		t.Fatalf("Short(stackBack=1) state некорректен: line=%d filename=%q", ti.Line, ti.FilenameLong.String())
	}
}

func TestShortNegativeStackBack(t *testing.T) {
	ti := kitTypes.NewTraceInfo()
	Short(ti, -2)
	if ti.Function.String() != "Short" {
		t.Fatalf("Function с отрицательным stackBack = %q", ti.Function.String())
	}
	if !strings.HasSuffix(ti.FilenameLong.String(), "trace.go") {
		t.Fatalf("FilenameLong с отрицательным stackBack = %q", ti.FilenameLong.String())
	}
}

func TestFunctionCorrect(t *testing.T) {
	var tests = []struct {
		name     string
		fn       string
		pkg      string
		function string
	}{
		{name: "simple", fn: "simple", pkg: "", function: "simple"},
		{name: "slash", fn: "pkg/sub.Func", pkg: "pkg/sub", function: "Func"},
		{name: "method", fn: "Type.Method", pkg: "Type", function: "Method"},
		{name: "slash-method", fn: "pkg/sub.Type.Method", pkg: "pkg/sub", function: "Type.Method"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ti := kitTypes.NewTraceInfo()
			_, _ = ti.Function.WriteString(test.fn)
			functionCorrect(ti)
			if ti.Package.String() != test.pkg || ti.Function.String() != test.function {
				t.Fatalf("functionCorrect(%q) = pkg:%q fn:%q", test.fn, ti.Package.String(), ti.Function.String())
			}
		})
	}
}

func TestGetStackBack(t *testing.T) {
	buf := getBuffer()
	defer putBuffer(buf)
	if n := getStackBack(buf); n < 2 {
		t.Fatalf("getStackBack() = %d, ожидалось минимум 2", n)
	}
}

func TestGetStackModes(t *testing.T) {
	var (
		tiFull = kitTypes.NewTraceInfo()
		buf    = getBuffer()
	)
	getStack(tiFull, 0, true, buf)
	putBuffer(buf)
	if !strings.Contains(tiFull.StackTrace.String(), "goroutine") {
		t.Fatalf("getStack(full) не вернул полный стек:\n%s", tiFull.StackTrace.String())
	}

	var tiNegative = kitTypes.NewTraceInfo()
	buf = getBuffer()
	getStack(tiNegative, -2, false, buf)
	putBuffer(buf)
	if !strings.Contains(tiNegative.StackTrace.String(), "goroutine") {
		t.Fatalf("getStack(negative stackBack) некорректен:\n%s", tiNegative.StackTrace.String())
	}

	var tiHuge = kitTypes.NewTraceInfo()
	buf = getBuffer()
	getStack(tiHuge, 100000, false, buf)
	putBuffer(buf)
	if tiHuge.StackTrace.Len() == 0 || !strings.HasPrefix(tiHuge.StackTrace.String(), "goroutine") {
		t.Fatalf("getStack(huge stackBack) некорректен:\n%s", tiHuge.StackTrace.String())
	}

	var tiGrowth = kitTypes.NewTraceInfo()
	buf = getBuffer()
	buf.Byte64k = make([]byte, 1)
	getStack(tiGrowth, 0, false, buf)
	putBuffer(buf)
	if tiGrowth.StackTrace.Len() == 0 || !strings.Contains(tiGrowth.StackTrace.String(), "goroutine") {
		t.Fatalf("getStack(small buffer) некорректен:\n%s", tiGrowth.StackTrace.String())
	}
}

func TestStackShortAndFull(t *testing.T) {
	short := StackShort()
	if len(short) == 0 || !strings.Contains(string(short), "goroutine") ||
		!strings.Contains(string(short), "TestStackShortAndFull") {
		t.Fatalf("StackShort() некорректен:\n%s", short)
	}
	full := StackFull()
	if len(full) == 0 || !strings.Contains(string(full), "goroutine") {
		t.Fatalf("StackFull() некорректен:\n%s", full)
	}
}

func TestStackGrowthSmallBuffer(t *testing.T) {
	short := stack(make([]byte, 1), false)
	if len(short) == 0 || !strings.Contains(string(short), "goroutine") {
		t.Fatalf("stack(small buffer, short) некорректен:\n%s", short)
	}
	full := stack(make([]byte, 1), true)
	if len(full) == 0 || !strings.Contains(string(full), "goroutine") {
		t.Fatalf("stack(small buffer, full) некорректен:\n%s", full)
	}
}

func BenchmarkShort(b *testing.B) {
	ti := kitTypes.NewTraceInfo()
	for i := 0; i < b.N; i++ {
		Short(ti, 0)
	}
}
