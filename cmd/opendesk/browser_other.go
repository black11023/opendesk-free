//go:build !windows

package main

import (
 "fmt"
 "os"
 "os/exec"
 "runtime"
)

func openSystemBrowser(url string) error {
 if runtime.GOOS=="darwin" {return exec.Command("open",url).Start()}
 return exec.Command("xdg-open",url).Start()
}
func showStartupError(err error) {fmt.Fprintln(os.Stderr,err)}
