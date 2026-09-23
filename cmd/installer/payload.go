//go:build !saylesspayload

package main

// bundledPayload is empty unless the installer is built with the
// `saylesspayload` build tag, which embeds sale.exe directly into say_less.exe.
var (
	bundledPayload      []byte
	bundledLogo         []byte
	bundledPortrait     []byte
	bundledFontRegular  []byte
	bundledFontSemibold []byte
)
