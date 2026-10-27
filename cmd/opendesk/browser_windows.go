//go:build windows

package main

import (
  "fmt"
  "syscall"
  "unsafe"
)

// Native Windows ShellExecute and MessageBox instead of launching
// rundll32.exe or a PowerShell interpreter during normal/error startup.
var shell32 = syscall.NewLazyDLL("shell32.dll")
var shellExecuteW = shell32.NewProc("ShellExecuteW")
var nativeUser32 = syscall.NewLazyDLL("user32.dll")
var nativeMessageBoxW = nativeUser32.NewProc("MessageBoxW")

func openSystemBrowser(url string) error {
  // Caller constructs this URL from a freshly bound 127.0.0.1 port only.
  if len(url)<18 || len(url)>64 || url[:17]!="http://127.0.0.1:" {
    return fmt.Errorf("refusing to launch a non-local URL")
  }
  action,_:=syscall.UTF16PtrFromString("open")
  target,err:=syscall.UTF16PtrFromString(url)
  if err!=nil { return err }
  rc,_,callErr:=shellExecuteW.Call(0,uintptr(unsafe.Pointer(action)),uintptr(unsafe.Pointer(target)),0,0,1)
  if rc<=32 {return fmt.Errorf("Windows failed to open local browser (code %d): %v",rc,callErr)}
  return nil
}

func showStartupError(err error) {
  title,_:=syscall.UTF16PtrFromString("OpenDesk Free")
  detail,_:=syscall.UTF16PtrFromString("OpenDesk Free cannot start.\n"+err.Error())
  nativeMessageBoxW.Call(0,uintptr(unsafe.Pointer(detail)),uintptr(unsafe.Pointer(title)),0x10)
}
