//go:build saylesspayload

package main

import _ "embed"

//go:embed sale.exe
var bundledPayload []byte

//go:embed say_less_logo.png
var bundledLogo []byte

//go:embed say_less_portrait_cover.png
var bundledPortrait []byte

//go:embed GoogleSansCodeNerdFontPropo-Regular.ttf
var bundledFontRegular []byte

//go:embed GoogleSansCodeNerdFontPropo-SemiBold.ttf
var bundledFontSemibold []byte
