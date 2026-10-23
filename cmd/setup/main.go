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

const version = "0.5.0-preview.1"

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
	if len(os.Args) > 1 && os.Args[1] == "/uninstall" {
		uninstall()
		return
	}
	if dialog("OpenDesk Free Setup", "Install OpenDesk Free v"+version+" for the current Windows user?\n\n100% offline. No administrator access needed.\n\nFiles will be installed in your Local AppData folder.\n\nDo you want to continue?", mbYesNo|mbInformation) != 6 {
		return
	}
	if err := install(); err != nil {
		dialog("Installation failed", err.Error(), mbWarning)
		return
	}
	if dialog("OpenDesk Free installed", "Installation complete.\n\nStart Menu → OpenDesk Free\n\nLaunch OpenDesk Free now?", mbYesNo|mbInformation) == 6 {
		root, e := installDir()
		if e == nil {
			_ = exec.Command(filepath.Join(root, "OpenDeskFree.exe")).Start()
		}
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
	script := `$ws=New-Object -ComObject WScript.Shell; $link=$ws.CreateShortcut(` + quotePS(link) + `); $link.TargetPath=` + quotePS(app) + `; $link.WorkingDirectory=` + quotePS(filepath.Dir(app)) + `; $link.Description='OpenDesk Free local-first workbench'; $link.Save()`
	return runPowershell(script)
}
func uninstall() {
	if dialog("Uninstall OpenDesk Free", "Remove OpenDesk Free from this PC?\n\nThis deletes only the application's installed files and Start Menu shortcut.\nDownloaded output files are kept.", mbYesNo|mbWarning) != 6 {
		return
	}
	root, err := installDir()
	if err != nil {
		dialog("Uninstall failed", err.Error(), mbWarning)
		return
	}
	app := filepath.Join(root, "OpenDeskFree.exe")
	if err = os.Remove(app); err != nil && !os.IsNotExist(err) {
		dialog("Close OpenDesk Free first", "Please close OpenDesk Free and run Uninstall again.\n\n"+err.Error(), mbWarning)
		return
	}
	_ = os.Remove(shortcutPath())
	key := `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\OpenDeskFree`
	_ = exec.Command("reg.exe", "delete", key, "/f").Run()
	self, err := os.Executable()
	if err == nil {
		// The uninstaller is running; after exit, remove its own binary and empty app directory.
		script := `Start-Sleep -Seconds 2; Remove-Item -LiteralPath ` + quotePS(self) + ` -Force -ErrorAction SilentlyContinue; Remove-Item -LiteralPath ` + quotePS(root) + ` -ErrorAction SilentlyContinue`
		code := utf16.Encode([]rune(script))
		raw := make([]byte, len(code)*2)
		for i, r := range code {
			raw[i*2] = byte(r)
			raw[i*2+1] = byte(r >> 8)
		}
		_ = exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-ExecutionPolicy", "Bypass", "-EncodedCommand", base64.StdEncoding.EncodeToString(raw)).Start()
	}
	dialog("OpenDesk Free", "Uninstalled. Your output documents were not changed.", mbInformation)
}
