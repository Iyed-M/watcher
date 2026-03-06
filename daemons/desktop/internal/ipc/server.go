package ipc

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"

	"github.com/Iyed-M/watcher/shared"
)

func StartListener(ctx context.Context, ch chan<- shared.WatcherEvent) {
	l, err := IPCListener()
	if err != nil {
		log.Fatal(err)
	}
	for {
		watcherConn, err := l.Accept()
		if err != nil {
			slog.Error("error accepting new watcher connection, skipping", "err", err)
			continue
		}
		go handleWatcher(watcherConn, ch)
	}
}

func handleWatcher(watcher io.ReadCloser, ch chan<- shared.WatcherEvent) {
	var event shared.WatcherEvent
	defer watcher.Close()

	sc := bufio.NewScanner(watcher)

	for sc.Scan() {
		if err := json.Unmarshal(sc.Bytes(), &event); err != nil {
			slog.Error("cant parse json event from watcher, ignoring event", "err", err)
		}
		ch <- event
	}
	err := sc.Err()
	if err != nil {
		slog.Error("error reading watcher message, closing watcher conn", "err", err)
	}
}
