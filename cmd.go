package fwatch

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

type Command struct {
	command []string  // Command to execute when some file is changed.
	cmd     *exec.Cmd // Current pointer to command.
	dir     string    // Directory to execute command
	pid     int       // Current pid of execution
	logger  *log.Logger
}

func (c *Command) Exec() error {
	if len(c.command) == 0 {
		return EmptyCommandErr
	}
	cmd := c.newCommand()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if c.cmd != nil {
		c.logger.Printf("Killing current execution of %v\n", c.cmd.Args)
		if err := syscall.Kill(-c.pid, syscall.SIGKILL); err != nil {
			return err
		}
	}
	c.cmd = cmd

	c.logger.Printf("Executing %v\n", c.cmd.Args)
	if err := c.cmd.Start(); err != nil {
		return err
	}

	c.pid = c.cmd.Process.Pid

	return nil
}

func (c *Command) newCommand() *exec.Cmd {
	cmd := exec.Command(c.command[0], c.command[1:]...)
	cmd.Dir = c.dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
