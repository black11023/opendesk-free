//go:build windows

package procrun

import(
 "context"
 "os"
 "os/exec"
 "strconv"
 "syscall"
 "time"
)

const createNewProcessGroup=0x00000200
func configureCancellation(cmd *exec.Cmd){
 cmd.SysProcAttr=&syscall.SysProcAttr{CreationFlags:createNewProcessGroup}
 cmd.Cancel=func()error{
  if cmd.Process==nil{return os.ErrProcessDone}
  // Windows CommandContext.Kill only kills the root. taskkill /T reaches
  // LibreOffice and OCR subprocesses, with a bounded fallback.
  ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second)
  defer cancel()
  if e:=exec.CommandContext(ctx,"taskkill.exe","/T","/F","/PID",strconv.Itoa(cmd.Process.Pid)).Run();e!=nil{
   return cmd.Process.Kill()
  }
  return nil
 }
}
