package server

import (
	"encoding/json"
	"net/http"

	"github.com/onemoreteam/httpframework"
	"github.com/onemoreteam/httpframework/modularity"
)

func init() { modularity.Register(&serverModule{}) }

type serverModule struct {
	modularity.Skeleton

	s *httpframework.Server
}

func (serverModule) Name() string { return "server" }

func (m *serverModule) Initialize(j json.RawMessage) (err error) {
	if err = json.Unmarshal(j, &cfg); err != nil {
		return
	}
	return
}

func (m *serverModule) Serve() (err error) {
	m.s = httpframework.New(
		&http.Server{
			Addr:    cfg.Listen,
			Handler: stdServeMux,
		},
		stdGrpcServerOptions...,
	)
	if err = m.s.RegisterGrpcService(stdGrpcServices...); err != nil {
		return
	}
	return m.s.ListenAndServe()
}

func (m *serverModule) Shutdown() {
	if m.s != nil {
		m.s.Close()
	}
}
