package modularity

import (
	"encoding/json"
)

var stdRegistry = newRegistry()

func Register(module Module) {
	if err := stdRegistry.Register(module); err != nil {
		panic(err)
	}
}

func Initialize(jb json.RawMessage) (err error) {
	return stdRegistry.Initialize(jb)
}
func Serve() {
	stdRegistry.Serve()
}
func Shutdown() {
	stdRegistry.Shutdown()
}
func Finalize() {
	stdRegistry.Finalize()
}
