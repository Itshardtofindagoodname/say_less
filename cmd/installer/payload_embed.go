//go:build saylesspayload

package main

import _ "embed"

//go:embed sale.exe
var bundledPayload []byte
