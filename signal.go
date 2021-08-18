package http_server_framework

import (
	"os"
	"os/signal"
)

func IgnoreSignal(sigs ...os.Signal) {
	signal.Ignore(sigs...)
}

type SignalWatcher struct {
	ch chan os.Signal
}

func WatchSignal(sigs ...os.Signal) SignalWatcher {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	return SignalWatcher{ch: ch}
}

func (sw SignalWatcher) Chan() <-chan os.Signal { return sw.ch }

func (sw SignalWatcher) Wait() os.Signal {
	return <-sw.ch
}
