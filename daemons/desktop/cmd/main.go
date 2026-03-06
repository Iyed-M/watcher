package main

import (
	"context"
	"sync"

	"github.com/Iyed-M/watcher/daemons/desktop/internal/aggregator"
	"github.com/Iyed-M/watcher/daemons/desktop/internal/ipc"
	"github.com/Iyed-M/watcher/shared"
)

func main() {
	ch := make(chan shared.WatcherEvent)
	ctx, cancel := context.WithCancel(context.Background())

	wg := new(sync.WaitGroup)

	go func() {
		wg.Add(1)
		defer wg.Done()
		ipc.StartListener(ctx, ch)
	}()

	go func() {
		wg.Add(1)
		go aggregator.StartAggregator(ctx, ch)
		defer wg.Done()
	}()
	_ = cancel

	wg.Wait()
}
