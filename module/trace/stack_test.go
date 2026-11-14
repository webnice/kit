package trace

import (
	"strings"
	"testing"
)

// TestStackShortAndFull Полный и короткий стек вызовов.
func TestStackShortAndFull(t *testing.T) {
	const (
		tplShortWrong = "StackShort() некорректен:\n%s."
		tplFullWrong  = "StackFull() некорректен:\n%s."
	)
	var (
		short []byte
		full  []byte
	)

	short = StackShort()
	if len(short) == 0 || !strings.Contains(string(short), "goroutine") ||
		!strings.Contains(string(short), "TestStackShortAndFull") {
		t.Fatalf(tplShortWrong, short)
	}
	full = StackFull()
	if len(full) == 0 || !strings.Contains(string(full), "goroutine") {
		t.Fatalf(tplFullWrong, full)
	}
}

// TestStackGrowthSmallBuffer Рост буфера при малом начальном размере.
func TestStackGrowthSmallBuffer(t *testing.T) {
	const (
		tplShortWrong = "stack(small buffer, short) некорректен:\n%s."
		tplFullWrong  = "stack(small buffer, full) некорректен:\n%s."
	)
	var (
		short []byte
		full  []byte
	)

	short = stack(make([]byte, 1), false)
	if len(short) == 0 || !strings.Contains(string(short), "goroutine") {
		t.Fatalf(tplShortWrong, short)
	}
	full = stack(make([]byte, 1), true)
	if len(full) == 0 || !strings.Contains(string(full), "goroutine") {
		t.Fatalf(tplFullWrong, full)
	}
}
