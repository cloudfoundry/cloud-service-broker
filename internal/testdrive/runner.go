package testdrive

import (
	"os/exec"
	"syscall"
	"time"
)

type runner struct {
	exited bool
	err    error
	cmd    *exec.Cmd
}

func newCommand(cmd *exec.Cmd) *runner {
	r := runner{cmd: cmd}
	if err := cmd.Start(); err != nil {
		return r.error(err)
	}
	go r.monitor()
	return &r
}

func (r *runner) error(err error) *runner {
	r.err = err
	r.exited = true
	return r
}

func (r *runner) gracefullStop() error {
	if r.exited {
		return nil
	}
	if r.cmd != nil && r.cmd.Process != nil {
		if err := r.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			return err
		}
	}

	for !r.exited {
		time.Sleep(time.Millisecond)
	}

	return nil
}

func (r *runner) forceStop() error {
	if r.exited {
		return nil
	}
	if r.cmd != nil && r.cmd.Process != nil {
		if err := r.cmd.Process.Signal(syscall.SIGKILL); err != nil {
			return err
		}
	}

	for !r.exited {
		time.Sleep(time.Millisecond)
	}

	return nil
}

// monitor waits for the command to exit and cleans up. It is typically run as a goroutine
func (r *runner) monitor() {
	r.err = r.cmd.Wait()
	r.exited = true
}
