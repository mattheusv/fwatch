package fwatch

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	dir           string   // Directory to watch files.
	ignorePattern []string // Pattern to ignore files when changed. E.g *_test
	pattern       []string // Pattern to watch files change.

	cmd     *command
	watcher *fsnotify.Watcher
	logger  *log.Logger
}

func NewWatcher(dir string, cmd, pattern, ignorePattern []string, logger *log.Logger) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		cmd: &command{
			command: cmd,
			dir:     dir,
			logger:  logger,
		},
		dir:           dir,
		pattern:       pattern,
		ignorePattern: ignorePattern,
		watcher:       watcher,
		logger:        logger,
	}, nil
}

func (w *Watcher) Watch() error {
	w.logger.Printf("Discovering sub directories in %s\n", w.dir)
	d, err := w.discoverSubDirectories(w.dir)
	if err != nil {
		return err
	}

	w.logger.Printf("Adding %d directories to watch\n", len(d))
	if err := w.addDirectories(d...); err != nil {
		return err
	}

	if err := w.cmd.Exec(); err != nil {
		return err
	}

	w.logger.Printf("Starting watching for changes...")
	for {
		if err := w.events(); err != nil {
			return err
		}
	}
}

func (w *Watcher) Stop() error {
	if err := w.watcher.Close(); err != nil {
		return err
	}
	return w.cmd.Stop()
}

func (w Watcher) events() error {
	select {
	case event, ok := <-w.watcher.Events:
		if !ok {
			return nil
		}
		if event.Op&fsnotify.Create == fsnotify.Create {
			newDirectories, err := w.discoverSubDirectories(event.Name)
			if err != nil {
				return err
			}
			w.logger.Printf("find new directories: %v\n", newDirectories)
			if err := w.addDirectories(newDirectories...); err != nil {
				return err
			}
			return nil
		}
		if event.Op&fsnotify.Write == fsnotify.Write {
			if err := w.cmd.Exec(); err != nil {
				return err
			}

		}
	case err, ok := <-w.watcher.Errors:
		if !ok {
			return fmt.Errorf("watcher files changes error: %v", err)
		}
	}
	return nil
}

func (w Watcher) addDirectories(directories ...string) error {
	for _, d := range directories {
		if err := w.watcher.Add(d); err != nil {
			return err
		}
	}
	return nil
}

func (w Watcher) isToIgnoreFile(file string) (bool, error) {
	for _, pattern := range w.ignorePattern {
		matched, err := filepath.Match(pattern, file)
		if err != nil {
			return true, err
		}
		if matched {
			return matched, nil
		}
	}
	return false, nil
}

func (w Watcher) discoverSubDirectories(dir string) ([]string, error) {
	directories := []string{}
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			directories = append(directories, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return directories, nil
}
