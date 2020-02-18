package watchfs

import (
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	golog "github.com/ipfs/go-log"
)

var log = golog.Logger("watchfs")

// EventPath stores information about a path that is capable of generating events
type EventPath struct {
	Path     string
	Username string
	Dsname   string
}

// FilesysWatcher will watch a set of directory paths, and send messages on a channel for events
// concerning those paths. These events are:
// * A new file in one of those paths was created
// * An existing file was modified / written to
// * An existing file was deleted
// * One of the folders being watched was renamed, but that folder is still being watched
// * One of the folders was removed, which makes it no longer watched
// TODO(dlong): Folder rename and remove are not supported yet.
type FilesysWatcher struct {
	Watcher *fsnotify.Watcher
	Sender  chan FilesysEvent
	Assoc   map[string]EventPath
}

// NewFilesysWatcher returns a new FilesysWatcher
func NewFilesysWatcher() *FilesysWatcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	return &FilesysWatcher{Watcher: watcher}
}

// Begin will start watching the given directory paths
func (w *FilesysWatcher) Begin(paths []EventPath) chan FilesysEvent {
	// Associate paths with additional information
	assoc := make(map[string]EventPath)

	for _, p := range paths {
		err := w.Watcher.Add(p.Path)
		if err != nil {
			log.Errorf("%s", err)
		}
		assoc[p.Path] = p
	}

	messages := make(chan FilesysEvent)
	w.Sender = messages
	w.Assoc = assoc

	// Dispatch filesystem events
	go func() {
		for {
			select {
			case event, ok := <-w.Watcher.Events:
				if !ok {
					log.Debugf("error getting event")
					continue
				}

				if event.Op == fsnotify.Chmod {
					// Don't care about CHMOD, skip it
					continue
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					w.sendEvent(ModifyFileEvent, event.Name, "")
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					w.sendEvent(CreateNewFileEvent, event.Name, "")
				}
			}
		}
	}()

	return messages
}

// sendEvent sends a message on the channel about an event
func (w *FilesysWatcher) sendEvent(etype EventType, sour, dest string) {
	dir := filepath.Dir(sour)
	ep := w.Assoc[dir]
	event := FilesysEvent{
		Type:        etype,
		Username:    ep.Username,
		Dsname:      ep.Dsname,
		Source:      sour,
		Destination: dest,
		Time:        time.Now(),
	}
	w.Sender <- event
}
