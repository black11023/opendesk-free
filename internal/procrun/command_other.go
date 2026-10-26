//go:build !linux && !darwin && !windows

package procrun

import "os/exec"

// An unsupported OS uses Go's default CommandContext process cancellation.
func configureCancellation(cmd *exec.Cmd){}
