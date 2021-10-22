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
func Deregister(names ...string) error {
	return stdRegistry.Deregister(names...)
}
func DeregisterAllExcept(names ...string) error {
	return stdRegistry.DeregisterAllExcept(names...)
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
