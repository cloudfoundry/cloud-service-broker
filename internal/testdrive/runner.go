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

// gracefulStop sends a SIGTERM and waits for the process to stop
// See also: requestStop(), forceStop()
func (r *runner) gracefulStop() error {
	if r.exited {
		return nil
	}
	if err := r.signal(syscall.SIGTERM); err != nil {
		return err
	}

	for !r.exited {
		time.Sleep(time.Millisecond)
	}

	return nil
}

// requestStop sends a SIGTERM and does not wait
// See also: gracefulStop(), forceStop()
func (r *runner) requestStop() error {
	if r.exited {
		return nil
	}
	if err := r.signal(syscall.SIGTERM); err != nil {
		return err
	}

	return nil
}

func (r *runner) forceStop() error {
	if r.exited {
		return nil
	}
	if err := r.signal(syscall.SIGKILL); err != nil {
		return err
	}

	for !r.exited {
		time.Sleep(time.Millisecond)
	}

	return nil
}

func (r *runner) signal(sig syscall.Signal) error {
	if r.cmd != nil && r.cmd.Process != nil {
		if err := r.cmd.Process.Signal(sig); err != nil {
			return err
		}
	}

	return nil
}

// monitor waits for the command to exit and cleans up. It is typically run as a goroutine
func (r *runner) monitor() {
	r.err = r.cmd.Wait()
	r.exited = true
}
