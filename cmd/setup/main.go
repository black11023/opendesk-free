//go:build windows

package main

import (
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

//go:embed payload/OpenDeskFree.exe
var payload embed.FS

const version = "0.6.0-preview.1"

var user32 = syscall.NewLazyDLL("user32.dll")
var msgBox = user32.NewProc("MessageBoxW")

func dialog(title, message string, flags uintptr) int {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(message)
	result, _, _ := msgBox.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), flags)
	return int(result)
}

const mbYesNo = 0x04
const mbInformation = 0x40
const mbWarning = 0x30

func main() {
 // Explicit unattended modes are useful for automated installation tests.
 // They never request elevation, never choose a custom installation path,
 // and remain scoped to the current Windows user's Local AppData.
 if len(os.Args)>1 {
   switch os.Args[1] {
   case "--quiet-install":
     if err:=install();err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
     return
   case "--quiet-uninstall":
     if err:=uninstall(true);err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
     return
   case "/uninstall":
     if err:=uninstall(false);err!=nil{dialog("Uninstall failed",err.Error(),mbWarning)}
     return
   }
 }
 if dialog("OpenDesk Free Setup","Install OpenDesk Free v"+version+" for the current Windows user?\n\nNo administrator access needed.\n\nFiles will be installed in your Local AppData folder.\n\nDo you want to continue?",mbYesNo|mbInformation)!=6{return}
 if err:=install();err!=nil{dialog("Installation failed",err.Error(),mbWarning);return}
 if dialog("OpenDesk Free installed","Installation complete.\n\nStart Menu → OpenDesk Free\n\nLaunch OpenDesk Free now?",mbYesNo|mbInformation)==6{
  if root,e:=installDir();e==nil{_ =exec.Command(filepath.Join(root,"OpenDeskFree.exe")).Start()}
 }
}
func installDir() (string, error) {
	p := os.Getenv("LOCALAPPDATA")
	if p == "" {
		return "", errors.New("LOCALAPPDATA is not defined")
	}
	return filepath.Join(p, "Programs", "OpenDeskFree"), nil
}
func install() error {
	root, err := installDir()
	if err != nil {
		return err
	}
    if os.Getenv("APPDATA")==""{return errors.New("APPDATA is not defined")}
	if err = os.MkdirAll(root, 0755); err != nil {
		return err
	}
	exeBytes, err := payload.ReadFile("payload/OpenDeskFree.exe")
	if err != nil {
		return err
	}
	// Always preserve existing app if its replacement cannot be written safely.
	app := filepath.Join(root, "OpenDeskFree.exe")
	tmp := app + ".new"
	if err = os.WriteFile(tmp, exeBytes, 0755); err != nil {
		return err
	}
	if err = os.Rename(tmp, app); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close any running OpenDesk Free process and run Setup again: %w", err)
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	src, err := os.ReadFile(self)
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(root, "OpenDeskFree_Uninstall.exe"), src, 0755); err != nil {
		return err
	}
	if err = createShortcut(app); err != nil {
		return fmt.Errorf("files installed but Start Menu shortcut failed: %w", err)
	}
	// Registry registration enables Settings > Installed apps, with per-user uninstall.
	key := `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenDeskFree`
	uninstall := `"` + filepath.Join(root, "OpenDeskFree_Uninstall.exe") + `" /uninstall`
	props := map[string]string{"DisplayName": "OpenDesk Free", "DisplayVersion": version, "Publisher": "OpenDesk Free Community", "InstallLocation": root, "UninstallString": uninstall, "DisplayIcon": app}
	for n, v := range props {
		if e := exec.Command("reg.exe", "add", key, "/v", n, "/t", "REG_SZ", "/d", v, "/f").Run(); e != nil {
			return fmt.Errorf("Windows registry failed (%s): %w", n, e)
		}
	}
	return nil
}
func shortcutPath() string {
	appdata := os.Getenv("APPDATA")
	return filepath.Join(appdata, "Microsoft", "Windows", "Start Menu", "Programs", "OpenDesk Free.lnk")
}
func quotePS(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
func runPowershell(script string) error {
	code := utf16.Encode([]rune(script))
	raw := make([]byte, len(code)*2)
	for i, r := range code {
		raw[i*2] = byte(r)
		raw[i*2+1] = byte(r >> 8)
	}
	encoded := base64.StdEncoding.EncodeToString(raw)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-EncodedCommand", encoded)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return fmt.Errorf("PowerShell failed: %w (%s)", e, string(out))
	}
	return nil
}
func createShortcut(app string) error {
	link := shortcutPath()
    // A newly created Windows user profile (and CI test profiles) may not
    // contain the Start Menu programs directory yet.
    if err:=os.MkdirAll(filepath.Dir(link),0755);err!=nil{return fmt.Errorf("create Start Menu directory: %w",err)}
	script := `$ws=New-Object -ComObject WScript.Shell; $link=$ws.CreateShortcut(` + quotePS(link) + `); $link.TargetPath=` + quotePS(app) + `; $link.WorkingDirectory=` + quotePS(filepath.Dir(app)) + `; $link.Description='OpenDesk Free local-first workbench'; $link.Save()`
	return runPowershell(script)
}
// uninstall removes only OpenDesk-owned executables, shortcut and the
// per-user registry entry. It never recursively deletes an installation
// directory, so user-created files in that folder are retained.
func uninstall(quiet bool) error {
 if !quiet&&dialog("Uninstall OpenDesk Free","Remove OpenDesk Free from this PC?\n\nThis deletes only installed application files and its Start Menu shortcut.\nDownloaded documents are kept.",mbYesNo|mbWarning)!=6{return nil}
 root,err:=installDir();if err!=nil{return err}
 if os.Getenv("APPDATA")==""{return errors.New("APPDATA is not defined")}
 app:=filepath.Join(root,"OpenDeskFree.exe")
 if err=os.Remove(app);err!=nil&&!os.IsNotExist(err){return fmt.Errorf("close OpenDesk Free before uninstalling: %w",err)}
 if err=os.Remove(shortcutPath());err!=nil&&!os.IsNotExist(err){return fmt.Errorf("remove Start Menu shortcut: %w",err)}
 key:=`HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenDeskFree`
 if err=exec.Command("reg.exe","delete",key,"/f").Run();err!=nil{
   // Repeating an uninstall with no registry key should be idempotent.
   if out,qerr:=exec.Command("reg.exe","query",key).CombinedOutput();qerr==nil{return fmt.Errorf("registry removal failed: %w",err)}else{_ =out}
 }
 self,err:=os.Executable()
 if err==nil {
   script:=`Start-Sleep -Seconds 2; Remove-Item -LiteralPath `+quotePS(self)+` -Force -ErrorAction SilentlyContinue; Remove-Item -LiteralPath `+quotePS(root)+` -ErrorAction SilentlyContinue`
   code:=utf16.Encode([]rune(script))
   raw:=make([]byte,len(code)*2)
   for i,v:=range code{raw[i*2]=byte(v);raw[i*2+1]=byte(v>>8)}
   if err=exec.Command("powershell.exe","-NoProfile","-NonInteractive","-WindowStyle","Hidden","-ExecutionPolicy","Bypass","-EncodedCommand",base64.StdEncoding.EncodeToString(raw)).Start();err!=nil{return fmt.Errorf("schedule uninstaller cleanup: %w",err)}
 }
 if !quiet{dialog("OpenDesk Free","Uninstalled. Output documents were not changed.",mbInformation)}
 return nil
}
