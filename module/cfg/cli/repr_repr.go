// Some values (such as pointers to basic types) can not be represented directly in
// Go. These values will be output as `&<value>`. eg. `&23`

package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"time"
	"unsafe"
)

var (
	// "Real" names of basic kinds, used to differentiate type aliases.
	realKindName = map[reflect.Kind]string{
		reflect.Bool:       "bool",
		reflect.Int:        "int",
		reflect.Int8:       "int8",
		reflect.Int16:      "int16",
		reflect.Int32:      "int32",
		reflect.Int64:      "int64",
		reflect.Uint:       "uint",
		reflect.Uint8:      "uint8",
		reflect.Uint16:     "uint16",
		reflect.Uint32:     "uint32",
		reflect.Uint64:     "uint64",
		reflect.Uintptr:    "uintptr",
		reflect.Float32:    "float32",
		reflect.Float64:    "float64",
		reflect.Complex64:  "complex64",
		reflect.Complex128: "complex128",
		reflect.Array:      "array",
		reflect.Chan:       "chan",
		reflect.Func:       "func",
		reflect.Map:        "map",
		reflect.Slice:      "slice",
		reflect.String:     "string",
	}

	goStringerType = reflect.TypeOf((*fmt.GoStringer)(nil)).Elem()

	byteSliceType = reflect.TypeOf([]byte{})
)

// ReprDefault prints to os.Stdout with two space indentation.
var ReprDefault = ReprNew(os.Stdout, ReprIndent("  "))

// An ReprOption modifies the default behaviour of a Printer.
type ReprOption func(o *ReprPrinter)

// ReprIndent output by this much.
func ReprIndent(indent string) ReprOption { return func(o *ReprPrinter) { o.indent = indent } }

// ReprNoIndent disables indenting.
func ReprNoIndent() ReprOption { return ReprIndent("") }

// ReprOmitEmpty sets whether empty field members should be omitted from output.
func ReprOmitEmpty(omitEmpty bool) ReprOption {
	return func(o *ReprPrinter) { o.omitEmpty = omitEmpty }
}

// ReprExplicitTypes adds explicit typing to slice and map struct values that would normally be inferred by Go.
func ReprExplicitTypes(ok bool) ReprOption { return func(o *ReprPrinter) { o.explicitTypes = true } }

// ReprIgnoreGoStringer disables use of the .GoString() method.
func ReprIgnoreGoStringer() ReprOption { return func(o *ReprPrinter) { o.ignoreGoStringer = true } }

// ReprHide excludes the given types from representation, instead just printing the name of the type.
func ReprHide(ts ...any) ReprOption {
	return func(o *ReprPrinter) {
		for _, t := range ts {
			rt := reflect.Indirect(reflect.ValueOf(t)).Type()
			o.exclude[rt] = true
		}
	}
}

// ReprAlwaysIncludeType always includes explicit type information for each item.
func ReprAlwaysIncludeType() ReprOption { return func(o *ReprPrinter) { o.alwaysIncludeType = true } }

// ReprPrinter represents structs in a printable manner.
type ReprPrinter struct {
	indent            string
	omitEmpty         bool
	ignoreGoStringer  bool
	alwaysIncludeType bool
	explicitTypes     bool
	exclude           map[reflect.Type]bool
	w                 io.Writer
}

// ReprNew creates a new Printer on w with the given Options.
func ReprNew(w io.Writer, options ...ReprOption) *ReprPrinter {
	p := &ReprPrinter{
		w:         w,
		indent:    "  ",
		omitEmpty: true,
		exclude:   map[reflect.Type]bool{},
	}
	for _, option := range options {
		option(p)
	}
	return p
}

func (p *ReprPrinter) nextIndent(indent string) string {
	if p.indent != "" {
		return indent + p.indent
	}
	return ""
}

func (p *ReprPrinter) thisIndent(indent string) string {
	if p.indent != "" {
		return indent
	}
	return ""
}

// Print the values.
func (p *ReprPrinter) Print(vs ...any) {
	for i, v := range vs {
		if i > 0 {
			fmt.Fprint(p.w, " ")
		}
		p.reprValue(map[reflect.Value]bool{}, reflect.ValueOf(v), "", true)
	}
}

// Println prints each value on a new line.
func (p *ReprPrinter) Println(vs ...any) {
	for i, v := range vs {
		if i > 0 {
			fmt.Fprint(p.w, " ")
		}
		p.reprValue(map[reflect.Value]bool{}, reflect.ValueOf(v), "", true)
	}
	fmt.Fprintln(p.w)
}

func (p *ReprPrinter) reprValue(seen map[reflect.Value]bool, v reflect.Value, indent string, showType bool) { // nolint: gocyclo
	if seen[v] {
		fmt.Fprint(p.w, "...")
		return
	}
	seen[v] = true
	defer delete(seen, v)

	if v.Kind() == reflect.Invalid || (v.Kind() == reflect.Ptr || v.Kind() == reflect.Map || v.Kind() == reflect.Chan || v.Kind() == reflect.Slice || v.Kind() == reflect.Func || v.Kind() == reflect.Interface) && v.IsNil() {
		fmt.Fprint(p.w, "nil")
		return
	}
	if p.exclude[v.Type()] {
		fmt.Fprintf(p.w, "%s...", v.Type().Name())
		return
	}
	t := v.Type()

	if t == byteSliceType {
		fmt.Fprintf(p.w, "[]byte(%q)", v.Bytes())
		return
	}

	// If we can't access a private field directly with reflection, try and do so via unsafe.
	if !v.CanInterface() && v.CanAddr() {
		uv := reflect.NewAt(t, unsafe.Pointer(v.UnsafeAddr())).Elem()
		if uv.CanInterface() {
			v = uv
		}
	}
	// Attempt to use fmt.GoStringer interface.
	if !p.ignoreGoStringer && t.Implements(goStringerType) {
		fmt.Fprint(p.w, v.Interface().(fmt.GoStringer).GoString())
		return
	}
	in := p.thisIndent(indent)
	ni := p.nextIndent(indent)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		fmt.Fprintf(p.w, "%s{", v.Type())
		if v.Len() == 0 {
			fmt.Fprint(p.w, "}")
		} else {
			if p.indent != "" {
				fmt.Fprintf(p.w, "\n")
			}
			for i := 0; i < v.Len(); i++ {
				e := v.Index(i)
				fmt.Fprintf(p.w, "%s", ni)
				p.reprValue(seen, e, ni, p.alwaysIncludeType || p.explicitTypes)
				if p.indent != "" {
					fmt.Fprintf(p.w, ",\n")
				} else if i < v.Len()-1 {
					fmt.Fprintf(p.w, ", ")
				}
			}
			fmt.Fprintf(p.w, "%s}", in)
		}

	case reflect.Chan:
		fmt.Fprintf(p.w, "make(")
		fmt.Fprintf(p.w, "%s", v.Type())
		fmt.Fprintf(p.w, ", %d)", v.Cap())

	case reflect.Map:
		fmt.Fprintf(p.w, "%s{", v.Type())
		if p.indent != "" && v.Len() != 0 {
			fmt.Fprintf(p.w, "\n")
		}
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j])
		})
		for i, k := range keys {
			kv := v.MapIndex(k)
			fmt.Fprintf(p.w, "%s", ni)
			p.reprValue(seen, k, ni, p.alwaysIncludeType || p.explicitTypes)
			fmt.Fprintf(p.w, ": ")
			p.reprValue(seen, kv, ni, true)
			if p.indent != "" {
				fmt.Fprintf(p.w, ",\n")
			} else if i < v.Len()-1 {
				fmt.Fprintf(p.w, ", ")
			}
		}
		fmt.Fprintf(p.w, "%s}", in)

	case reflect.Struct:
		if td, ok := v.Interface().(time.Time); ok {
			timeToGo(p.w, td)
		} else {
			if showType {
				fmt.Fprintf(p.w, "%s{", v.Type())
			} else {
				fmt.Fprint(p.w, "{")
			}
			if p.indent != "" && v.NumField() != 0 {
				fmt.Fprintf(p.w, "\n")
			}
			for i := 0; i < v.NumField(); i++ {
				t := v.Type().Field(i)
				f := v.Field(i)
				if p.omitEmpty && isZero(f) {
					continue
				}
				fmt.Fprintf(p.w, "%s%s: ", ni, t.Name)
				p.reprValue(seen, f, ni, true)
				if p.indent != "" {
					fmt.Fprintf(p.w, ",\n")
				} else if i < v.NumField()-1 {
					fmt.Fprintf(p.w, ", ")
				}
			}
			fmt.Fprintf(p.w, "%s}", indent)
		}
	case reflect.Ptr:
		if v.IsNil() {
			fmt.Fprintf(p.w, "nil")
			return
		}
		if showType {
			fmt.Fprintf(p.w, "&")
		}
		p.reprValue(seen, v.Elem(), indent, showType)

	case reflect.String:
		if t.Name() != "string" || p.alwaysIncludeType {
			fmt.Fprintf(p.w, "%s(%q)", t, v.String())
		} else {
			fmt.Fprintf(p.w, "%q", v.String())
		}

	case reflect.Interface:
		if v.IsNil() {
			fmt.Fprintf(p.w, "interface {}(nil)")
		} else {
			p.reprValue(seen, v.Elem(), indent, true)
		}

	default:
		if t.Name() != realKindName[t.Kind()] || p.alwaysIncludeType {
			fmt.Fprintf(p.w, "%s(%v)", t, v)
		} else {
			fmt.Fprintf(p.w, "%v", v)
		}
	}
}

// ReprString returns a string representing v.
func ReprString(v any, options ...ReprOption) string {
	w := bytes.NewBuffer(nil)
	options = append([]ReprOption{ReprNoIndent()}, options...)
	p := ReprNew(w, options...)
	p.Print(v)
	return w.String()
}

func extractOptions(vs ...any) (args []any, options []ReprOption) {
	for _, v := range vs {
		if o, ok := v.(ReprOption); ok {
			options = append(options, o)
		} else {
			args = append(args, v)
		}
	}
	return
}

// ReprPrintln prints v to os.Stdout, one per line.
func ReprPrintln(vs ...any) {
	args, options := extractOptions(vs...)
	ReprNew(os.Stdout, options...).Println(args...)
}

// ReprPrint writes a representation of v to os.Stdout, separated by spaces.
func ReprPrint(vs ...any) {
	args, options := extractOptions(vs...)
	ReprNew(os.Stdout, options...).Print(args...)
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func timeToGo(w io.Writer, t time.Time) {
	if t.IsZero() {
		fmt.Fprint(w, "time.Time{}")
		return
	}

	var zone string
	switch loc := t.Location(); loc {
	case nil:
		zone = "nil"
	case time.UTC:
		zone = "time.UTC"
	case time.Local:
		zone = "time.Local"
	default:
		n, off := t.Zone()
		zone = fmt.Sprintf("time.FixedZone(%q, %d)", n, off)
	}
	y, m, d := t.Date()
	fmt.Fprintf(w, `time.Date(%d, %d, %d, %d, %d, %d, %d, %s)`, y, m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), zone)
}
