package fwatch

import (
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

const EmptyCommandErr = FsWatchError("Could execute empty command")
const WatcherClosedError = FsWatchError("Watcher closed")

type FsWatchError string

func (err FsWatchError) Error() string {
	return string(err)
}

type command struct {
	command []string  // Command to execute when some file is changed.
	cmd     *exec.Cmd // Current pointer to command.
	dir     string    // Directory to execute command
	pid     int       // Current pid of execution
	logger  *log.Logger
	mux     sync.Mutex
}

func (c *command) Exec() error {
	if len(c.command) == 0 {
		return EmptyCommandErr
	}
	cmd := c.newCommand()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := c.Stop(); err != nil {
		return err
	}

	c.mux.Lock()
	c.cmd = cmd

	c.logger.Printf("Executing %v\n", c.cmd.Args)
	if err := c.cmd.Start(); err != nil {
		return err
	}

	c.pid = c.cmd.Process.Pid
	c.mux.Unlock()

	return nil
}

func (c *command) Stop() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.cmd != nil {
		c.logger.Printf("Killing current execution of %v\n", c.cmd.Args)
		return syscall.Kill(-c.pid, syscall.SIGKILL)
	}
	return nil
}

func (c *command) newCommand() *exec.Cmd {
	cmd := exec.Command(c.command[0], c.command[1:]...)
	cmd.Dir = c.dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
