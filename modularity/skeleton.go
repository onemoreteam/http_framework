package modularity

import (
	"encoding/json"
	"errors"
)

var notImplementedError = errors.New("not implemented")

type Skeleton struct{}

func (Skeleton) Priority() int { return 0 }

func (Skeleton) Initialize(json.RawMessage) error { return nil }

func (m Skeleton) Serve() error { return notImplementedError }

func (m Skeleton) Shutdown() {}

func (m Skeleton) Finalize() {}

/// make sure Skeleton implements Module except Name method

var _ Module = (*__skeleton__)(nil)

type __skeleton__ struct{ Skeleton }

func (__skeleton__) Name() string { return "" }
