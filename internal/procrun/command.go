// Package procrun starts optional local document tools without a shell.
// When the request is cancelled, it attempts to terminate the whole child
// process tree instead of leaving OCR/Office workers behind.
package procrun

import (
 "context"
 "os/exec"
 "time"
)

func Command(ctx context.Context,binary string,args ...string)*exec.Cmd{
 cmd:=exec.CommandContext(ctx,binary,args...)
 configureCancellation(cmd)
 cmd.WaitDelay=3*time.Second
 return cmd
}
