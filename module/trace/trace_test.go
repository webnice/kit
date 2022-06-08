// Package trace
package trace

import (
	"strings"
	"testing"

	kitTypes "github.com/webnice/kit/types"
)

func TestShort(t *testing.T) {
	ti := kitTypes.NewTraceInfo()
	Short(ti, 0)
	if ti.StackTrace.Len() == 0 {
		t.Fatalf("функция Short() работает не корректно")
		return
	}
	if strings.Contains(ti.StackTrace.String(), "trace_test.go") {
		t.Fatalf("функция Short() не обрезала стек")
		return
	}
}

func BenchmarkShort(b *testing.B) {
	ti := kitTypes.NewTraceInfo()
	for i := 0; i < b.N; i++ {
		Short(ti, 0)
	}
}
