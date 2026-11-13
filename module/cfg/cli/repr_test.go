package cli

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func equal(t *testing.T, want, have string) {
	t.Helper()
	if want != have {
		t.Errorf("\nWant: %q\nHave: %q", want, have)
	}
}

type anotherStruct struct {
	A []int
}

type testStruct struct {
	S string
	I *int
	A anotherStruct
}

func TestReprEmptyArray(t *testing.T) {
	equal(t, "[]string{}", ReprString([]string{}, ReprOmitEmpty(false)))
}

func TestReprStringArray(t *testing.T) {
	equal(t, "[]string{\"a\", \"b\"}", ReprString([]string{"a", "b"}))
}

func TestReprIntArray(t *testing.T) {
	equal(t, "[]int{1, 2}", ReprString([]int{1, 2}))
}

func TestReprPointerToInt(t *testing.T) {
	pi := new(int)
	*pi = 13
	equal(t, `&13`, ReprString(pi))
}

func TestReprChannel(t *testing.T) {
	ch := make(<-chan map[string]*testStruct, 1)
	equal(t, `make(<-chan map[string]*cli.testStruct, 1)`, ReprString(ch))
}

func TestReprEmptyMap(t *testing.T) {
	equal(t, "map[string]bool{}", ReprString(map[string]bool{}))
}

func TestReprMap(t *testing.T) {
	m := map[string]int{"b": 3, "a": 1, "c": 5}
	for i := 0; i < 1000; i++ {
		equal(t, "map[string]int{\"a\": 1, \"b\": 3, \"c\": 5}", ReprString(m))
	}
}

func TestReprIntMap(t *testing.T) {
	m := map[int]string{3: "b", 1: "a", 5: "c"}
	for i := 0; i < 1000; i++ {
		equal(t, "map[int]string{1: \"a\", 3: \"b\", 5: \"c\"}", ReprString(m))
	}
}

func TestReprStructWithReprIndent(t *testing.T) {
	pi := new(int)
	*pi = 13
	s := &testStruct{
		S: "String",
		I: pi,
		A: anotherStruct{
			A: []int{1, 2, 3},
		},
	}
	equal(t, `&cli.testStruct{
  S: "String",
  I: &13,
  A: cli.anotherStruct{
    A: []int{
      1,
      2,
      3,
    },
  },
}`, ReprString(s, ReprIndent("  ")))
}

func TestReprByteArray(t *testing.T) {
	b := []byte{1, 2, 3}
	equal(t, "[]byte(\"\\x01\\x02\\x03\")", ReprString(b))
}

type privateTestStruct struct {
	a string
}

func TestReprPrivateField(t *testing.T) {
	s := privateTestStruct{"hello"}
	equal(t, `cli.privateTestStruct{a: "hello"}`, ReprString(s))
}

func TestReprNilAlone(t *testing.T) {
	var err error
	s := ReprString(err)
	equal(t, "nil", s)
}

func TestReprExplicitTypes(t *testing.T) {
	arr := []*privateTestStruct{{"hello"}, nil}
	s := ReprString(arr, ReprExplicitTypes(true))
	equal(t, "[]*cli.privateTestStruct{&cli.privateTestStruct{a: \"hello\"}, nil}", s)
}

func TestReprNilInsideArray(t *testing.T) {
	arr := []*privateTestStruct{{"hello"}, nil}
	s := ReprString(arr)
	equal(t, "[]*cli.privateTestStruct{{a: \"hello\"}, nil}", s)
}

func TestReprEmptySlice(t *testing.T) {
	a := []int{}
	s := ReprString(a)
	equal(t, "[]int{}", s)
}

func TestReprNilSlice(t *testing.T) {
	var a []int
	s := ReprString(a)
	equal(t, "nil", s)
}

type intSliceStruct struct{ f []int }

func TestReprEmptySliceStruct(t *testing.T) {
	a := intSliceStruct{f: []int{}}
	s := ReprString(a)
	equal(t, "cli.intSliceStruct{f: []int{}}", s)
}

func TestReprNilSliceStruct(t *testing.T) {
	var a intSliceStruct
	s := ReprString(a)
	equal(t, "cli.intSliceStruct{}", s)
}

type Enum int

func (e Enum) String() string {
	return "Value"
}

func TestEnum(t *testing.T) {
	v := Enum(1)
	s := ReprString(v)
	equal(t, "cli.Enum(Value)", s)
}

func TestShowType(t *testing.T) {
	a := map[string]privateTestStruct{"foo": {"bar"}}
	s := ReprString(a, ReprAlwaysIncludeType(), ReprIndent("  "))
	equal(t, strings.TrimSpace(`
map[string]cli.privateTestStruct{
  string("foo"): cli.privateTestStruct{
    a: string("bar"),
  },
}
`), s)
}

func TestRecursiveIssue3(t *testing.T) {
	type data struct {
		parent   *data
		children []*data
	}
	child := &data{}
	root := &data{children: []*data{child}}
	child.parent = root
	want := "&cli.data{children: []*cli.data{{parent: &..., }}}"
	have := ReprString(root)
	equal(t, want, have)
}

type MyBuffer struct {
	buf *bytes.Buffer
}

func TestReprPrivateBytes(t *testing.T) {
	mb := MyBuffer{
		buf: bytes.NewBufferString("Hi th3re!"),
	}
	s := ReprString(mb)

	switch v := runtime.Version(); {
	case strings.Contains(v, "go1.9"):
		equal(t, "cli.MyBuffer{buf: &bytes.Buffer{buf: []byte(\"Hi th3re!\"), bootstrap: [64]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}}", s)
	case strings.Contains(v, "go1.10"), strings.Contains(v, "go1.11"):
		equal(t, "cli.MyBuffer{buf: &bytes.Buffer{buf: []byte(\"Hi th3re!\"), bootstrap: [64]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, }}", s)
	default:
		equal(t, "cli.MyBuffer{buf: &bytes.Buffer{buf: []byte(\"Hi th3re!\"), }}", s)
	}
}
