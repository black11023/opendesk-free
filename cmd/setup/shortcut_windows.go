//go:build windows

package main

import (
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "syscall"
    "unsafe"
)

// Windows-native Start Menu shortcut, via COM IShellLinkW/IPersistFile.
// Avoids PowerShell -EncodedCommand / ExecutionPolicy Bypass, requires no
// elevation and supports Unicode paths. Based on documented Shell APIs.
type guid struct {
    Data1 uint32
    Data2 uint16
    Data3 uint16
    Data4 [8]byte
}
var (
    ole32 = syscall.NewLazyDLL("ole32.dll")
    coInitializeEx = ole32.NewProc("CoInitializeEx")
    coUninitialize = ole32.NewProc("CoUninitialize")
    coCreateInstance = ole32.NewProc("CoCreateInstance")
    clsidShellLink = guid{Data1:0x00021401,Data4:[8]byte{0xc0,0,0,0,0,0,0,0x46}}
    iidShellLinkW = guid{Data1:0x000214f9,Data4:[8]byte{0xc0,0,0,0,0,0,0,0x46}}
    iidPersistFile = guid{Data1:0x0000010b,Data4:[8]byte{0xc0,0,0,0,0,0,0,0x46}}
)
func comHR(op string, h uintptr) error {
    if int32(uint32(h))<0 { return fmt.Errorf("%s: HRESULT 0x%08x",op,uint32(h)) }
    return nil
}
func comCall(instance unsafe.Pointer,method uintptr,args ...uintptr) uintptr {
    vtable:=*(*unsafe.Pointer)(instance)
    address:=*(*uintptr)(unsafe.Add(vtable,method*unsafe.Sizeof(uintptr(0))))
    arguments:=append([]uintptr{uintptr(instance)},args...)
    result,_,_:=syscall.SyscallN(address,arguments...)
    return result
}
func createNativeShortcut(target,link string) error {
    if err:=os.MkdirAll(filepath.Dir(link),0755);err!=nil{return fmt.Errorf("create Start Menu directory: %w",err)}
    runtime.LockOSThread()
    defer runtime.UnlockOSThread()
    h,_,_:=coInitializeEx.Call(0,2) // COINIT_APARTMENTTHREADED
    if err:=comHR("CoInitializeEx",h);err!=nil{return err}
    defer coUninitialize.Call()
    var instance unsafe.Pointer
    h,_,_=coCreateInstance.Call(uintptr(unsafe.Pointer(&clsidShellLink)),0,1,uintptr(unsafe.Pointer(&iidShellLinkW)),uintptr(unsafe.Pointer(&instance)))
    if err:=comHR("CoCreateInstance ShellLink",h);err!=nil{return err}
    if instance==nil{return fmt.Errorf("CoCreateInstance returned a nil ShellLink")}
    defer comCall(instance,2)
    targetW,e:=syscall.UTF16PtrFromString(target);if e!=nil{return e}
    if e=comHR("IShellLinkW.SetPath",comCall(instance,20,uintptr(unsafe.Pointer(targetW))));e!=nil{return e}
    workingW,e:=syscall.UTF16PtrFromString(filepath.Dir(target));if e!=nil{return e}
    if e=comHR("IShellLinkW.SetWorkingDirectory",comCall(instance,9,uintptr(unsafe.Pointer(workingW))));e!=nil{return e}
    descW,e:=syscall.UTF16PtrFromString("OpenDesk Free local document tools");if e!=nil{return e}
    if e=comHR("IShellLinkW.SetDescription",comCall(instance,7,uintptr(unsafe.Pointer(descW))));e!=nil{return e}
    var persist unsafe.Pointer
    if e=comHR("IShellLinkW.QueryInterface",comCall(instance,0,uintptr(unsafe.Pointer(&iidPersistFile)),uintptr(unsafe.Pointer(&persist))));e!=nil{return e}
    if persist==nil{return fmt.Errorf("ShellLink did not implement IPersistFile")}
    defer comCall(persist,2)
    linkW,e:=syscall.UTF16PtrFromString(link);if e!=nil{return e}
    if e=comHR("IPersistFile.Save",comCall(persist,6,uintptr(unsafe.Pointer(linkW)),1));e!=nil{return e}
    return nil
}
