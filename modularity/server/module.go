package server

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/onemoreteam/httpframework/modularity"
)

func init() { modularity.Register(&serverModule{}) }

type serverModule struct {
	modularity.Skeleton

	lis net.Listener
}

func (serverModule) Name() string { return "server" }

func (m *serverModule) Initialize(j json.RawMessage) (err error) {
	if err = json.Unmarshal(j, &cfg); err != nil {
		return
	}
	m.lis, err = net.Listen("tcp", cfg.Listen)
	if err != nil {
		return
	}
	return
}

func (m *serverModule) Finalize() {
	if m.lis != nil {
		m.lis.Close()
		m.lis = nil
	}
}

func (m *serverModule) Serve() (err error) {
	return stdServer.Serve(m.lis)
}

func (m *serverModule) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stdServer.Shutdown(ctx)
}
