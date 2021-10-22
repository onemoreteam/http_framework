package modularity

import (
	"encoding/json"
	"errors"
)

var notImplementedError = errors.New("not implemented")

type Skeleton struct {
}

func (Skeleton) Priority() int { return 0 }

func (Skeleton) Initialize(json.RawMessage) error { return nil }

func (m Skeleton) Serve() error { return notImplementedError }

func (m Skeleton) Shutdown() {}

func (m Skeleton) Finalize() {}

////////////////////////////////////////////////////////////////////////////////

var _ Module = (*skeletonModule)(nil)

type skeletonModule struct {
	Skeleton
}

func (skeletonModule) Name() string  { return "" }
func (skeletonModule) Priority() int { return 0 }
