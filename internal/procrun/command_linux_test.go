//go:build linux

package procrun

import (
 "context"
 "errors"
 "os"
 "path/filepath"
 "strconv"
 "strings"
 "testing"
 "time"
)

func isRunning(pid int)bool{
 b,e:=os.ReadFile("/proc/"+strconv.Itoa(pid)+"/stat")
 if os.IsNotExist(e){return false}
 if e!=nil{return true}
 parts:=strings.SplitAfterN(string(b),") ",2)
 if len(parts)<2{return true}
 return !strings.HasPrefix(parts[1],"Z")
}
func TestCancellationKillsDescendantProcess(t *testing.T){
 ctx,cancel:=context.WithCancel(context.Background())
 defer cancel()
 marker:=filepath.Join(t.TempDir(),"child-pid.txt")
 // The background sleep must not outlive the shell process when cancelled.
 cmd:=Command(ctx,"sh","-c", "sleep 30 & echo $! > \""+marker+"\"; wait")
 if e:=cmd.Start();e!=nil{t.Fatal(e)}
 defer func(){cancel();_ =cmd.Process.Kill();_ =cmd.Wait()}()
 var pid int
 deadline:=time.Now().Add(3*time.Second)
 for time.Now().Before(deadline){
  b,e:=os.ReadFile(marker)
  if e==nil{
   pid,e=strconv.Atoi(strings.TrimSpace(string(b)))
   if e!=nil{t.Fatal(e)}
   break
  }
  time.Sleep(20*time.Millisecond)
 }
 if pid<1{t.Fatal("spawned descendant PID not recorded")}
 cancel()
 wait:=make(chan error,1);go func(){wait<-cmd.Wait()}()
 select{
 case e:=<-wait:
  if e==nil{t.Fatal("cancelled process exited successfully")}
 case <-time.After(5*time.Second):t.Fatal("cancelled process did not exit")
 }
 deadline=time.Now().Add(3*time.Second)
 for time.Now().Before(deadline){
  if !isRunning(pid){return}
  time.Sleep(25*time.Millisecond)
 }
 // Never let a leaked child remain after a failed regression test.
 p,_:=os.FindProcess(pid);if p!=nil{_ =p.Kill()}
 t.Fatalf("descendant %d remained running after cancellation",pid)
}
func TestCancelledContextRejectsLaunch(t *testing.T){
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 if e:=Command(ctx,"sh","-c","true").Run();!errors.Is(e,context.Canceled){t.Fatalf("unexpected error %v",e)}
}
