//go:build !windows || !amd64

package main

import "fmt"

func main() {
	fmt.Println("The Say Less GUI installer (say_less.exe) is only available for Windows x64.")
	fmt.Println("Use `go build -tags saylesspayload -o say_less.exe ./cmd/installer` on Windows.")
}
