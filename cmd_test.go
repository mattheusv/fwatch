package fwatch

import (
	"log"
	"testing"

	"github.com/tj/assert"
)

type FakeOutput struct {
	output []string
}

func (f *FakeOutput) Write(p []byte) (n int, err error) {
	f.output = append(f.output, string(p))
	return 0, nil
}

var fakeOut = FakeOutput{}

func TestExecRestartCommand(t *testing.T) {
	c := command{
		command: []string{"sleep", "0.5"},
		logger:  log.New(&fakeOut, "", log.Lmsgprefix),
	}

	err := c.Exec()
	assert.Nil(t, err, "Expected nil error on first execution")

	err = c.Exec()
	assert.Nil(t, err, "Expected nil error on second execution")
	assert.Contains(t, fakeOut.output, "Killing current execution of [sleep 0.5]\n", "Expected output of killing execution")
}

func TestExecSucessfull(t *testing.T) {
	c := command{
		command: []string{"echo", "===TEST==="},
		logger:  log.New(&fakeOut, "", log.Lmsgprefix),
	}

	err := c.Exec()

	assert.Nil(t, err, "Expected nil error")
}

func TestExecEmptyCommand(t *testing.T) {
	c := command{}

	err := c.Exec()

	assert.NotNil(t, err, "Expected error")
	assert.Equal(t, err, EmptyCommandErr)
}
