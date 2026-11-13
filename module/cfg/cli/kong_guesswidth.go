//go:build appengine || (!linux && !freebsd && !darwin && !dragonfly && !netbsd && !openbsd)
// +build appengine !linux,!freebsd,!darwin,!dragonfly,!netbsd,!openbsd

package cli

import "io"

func guessWidth(w io.Writer) int {
	return 80
}
