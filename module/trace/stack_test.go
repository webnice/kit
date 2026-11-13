package trace

import (
	"strings"
	"testing"
)

// TestStackShortAndFull Полный и короткий стек вызовов.
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

// TestStackGrowthSmallBuffer Рост буфера при малом начальном размере.
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
