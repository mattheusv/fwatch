package fwatch

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/tj/assert"
)

var (
	testCommand       = []string{"echo", "testing command"}
	testPattern       = []string{"*.go"}
	testIgnorePattern = []string{"*.js"}
	testLogger        *log.Logger
	verbose           = flag.Bool("V", false, "Enable logs in tests")
)

func TestMain(m *testing.M) {
	flag.Parse()
	if *verbose {
		testLogger = log.New(os.Stderr, "", log.Lshortfile)
	} else {
		testLogger = log.New(ioutil.Discard, "", log.Lshortfile)
	}
	os.Exit(m.Run())
}

func TestWatchEvents(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestWatchEventsDir")
	assert.Nil(t, err, "Expected nil error to create temp dir: %v", err)

	tmpFile, err := ioutil.TempFile(tmpDir, "TestWatchEventsFile")
	assert.Nil(t, err, "Expected nil error to create temp file: %v", err)
	defer tmpFile.Close()

	w, err := NewWatcher(tmpDir, testCommand, testPattern, testIgnorePattern, testLogger)
	assert.Nil(t, err, "Expected nil error to create watcher: %v", err)

	chErr := make(chan error)
	go func() {
		chErr <- w.Watch()
	}()

	time.Sleep(1 * time.Second) // TODO Find a better way
	for i := 0; i < 2; i++ {
		_, err := tmpFile.WriteString(fmt.Sprintf("Test %d", i))
		assert.Nil(t, err, "Expected nil error to write to temp file: %v", err)

		_, err = ioutil.TempDir(tmpDir, fmt.Sprintf("Test Dir %d", i))
		assert.Nil(t, err, "Expected nil error to create temp dir: %v", err)
	}

	time.Sleep(1 * time.Second) // TODO Find a better way
	err = w.Stop()
	assert.Nil(t, err, "Expected nil error to stop: %v", err)

	err = <-chErr
	assert.Nil(t, err, "Expected nil error: %v", err)
}

func TestWatchWithNilError(t *testing.T) {
	w, err := NewWatcher(".", testCommand, testPattern, testIgnorePattern, testLogger)

	assert.Nil(t, err, "Expected nil error to create watcher: %v", err)

	chErr := make(chan error)

	go func() {
		chErr <- w.Watch()
	}()

	time.Sleep(1 * time.Second) // TODO Find a better way

	err = w.Stop()
	assert.Nil(t, err, "Expected nil error to stop: %v", err)

	err = <-chErr
	assert.Nil(t, err, "Expected nil error: %v", err)
}

func TestNewWatcher(t *testing.T) {
	w, err := NewWatcher(".", testCommand, testPattern, testIgnorePattern, testLogger)

	assert.Nil(t, err, "Expected nil error to create watcher: %v", err)
	assert.NotNil(t, w.watcher, "Expected pointer to notify watcher")
	assert.NotNil(t, w.logger, "Expected pointer to logger")
	assert.NotNil(t, w.cmd.logger, "Expected pointer to cmd.logger")
	assert.Equal(t, w.cmd.command, testCommand, "Expected equal commands")
	assert.Equal(t, w.dir, w.dir, "Expected equal directories")
}
