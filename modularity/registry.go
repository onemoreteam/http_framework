package modularity

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	log "github.com/ntons/log-go"
)

type registry struct {
	mu      sync.Mutex // protect registry
	modules []Module
}

func newRegistry() *registry {
	return &registry{}
}

func (r *registry) Register(m Module) (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	i := sort.Search(len(r.modules), func(i int) bool {
		return r.modules[i].Name() >= m.Name()
	})
	if i < len(r.modules) && r.modules[i].Name() == m.Name() {
		return fmt.Errorf("module name %s had been registered", m.Name())
	}
	r.modules = append(r.modules, nil)
	copy(r.modules[i+1:], r.modules[i:])
	r.modules[i] = m
	return
}

func (r *registry) Initialize(j json.RawMessage) (err error) {
	var jm map[string]json.RawMessage
	if err = json.Unmarshal(j, &jm); err != nil {
		return
	}

	modules := append([]Module{}, r.modules...)
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Priority() < modules[j].Priority()
	})

	for i, module := range modules {
		if err = module.Initialize(jm[module.Name()]); err != nil {
			log.Warnf("failed to initalize module %s: %v", module.Name(), err)
			for j := i - 1; j >= 0; j-- {
				modules[j].Finalize()
			}
			return
		}
		log.Infof("module %s initialized", module.Name())
	}

	return
}

func (r *registry) Serve() {
	var wg sync.WaitGroup
	defer wg.Wait()

	for _, m := range r.modules {
		wg.Add(1)
		go func(m Module) {
			defer wg.Done()
			if err := m.Serve(); err != nil {
				if err != notImplementedError {
					log.Infof("module %s exited with error %w", m.Name(), err)
				}
			} else {
				log.Infof("module %s exited gracefully", m.Name())
			}
		}(m)
	}
	wg.Wait()
	return
}

func (r *registry) Shutdown() {
	for _, m := range r.modules {
		m.Shutdown()
	}
}

func (r *registry) Finalize() {
	for _, m := range r.modules {
		m.Finalize()
	}
}
