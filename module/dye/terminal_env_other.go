//go:build js || plan9 || aix
// +build js plan9 aix

package dye

func colorProfile() terminalProfile {
	return terminalANSI256
}

func foregroundColor() Color {
	return colorAnsi(7)
}

func backgroundColor() Color {
	return colorAnsi(0)
}
