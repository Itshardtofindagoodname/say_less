//go:build !windows

package main

import "fmt"

// cmdUninstall is only supported on Windows, where install.bat registers the
// CLI in the user's PATH. On other platforms the binary is a plain drop-in and
// uninstalling just means deleting the file.
func cmdUninstall() {
	fmt.Println("sale -g uninstall is only supported on Windows.")
	fmt.Println("On this platform Say Less is a single file - delete the 'sale' binary to uninstall.")
}
