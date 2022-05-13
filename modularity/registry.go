package modularity

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	log "github.com/ntons/log-go"
)

type registry struct {
	mu          sync.Mutex // protect registry
	modules     []Module
	initialized bool
}

func newRegistry() *registry {
	return &registry{}
}

func (r *registry) Register(m Module) (err error) {
	if r.initialized {
		return fmt.Errorf("registration was closed")
	}
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

func (r *registry) Deregister(names ...string) error {
	if r.initialized {
		return fmt.Errorf("registration was closed")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, name := range names {
		if i := sort.Search(len(r.modules), func(i int) bool {
			return r.modules[i].Name() >= name
		}); i < len(r.modules) && r.modules[i].Name() == name {
			r.modules = append(r.modules[:i], r.modules[i+1:]...)
		} else {
			return fmt.Errorf("unregistered module %s", name)
		}
	}
	return nil
}

func (r *registry) DeregisterAllExcept(names ...string) error {
	if r.initialized {
		return fmt.Errorf("registration was closed")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	modules := make([]Module, 0)
	for _, name := range names {
		if i := sort.Search(len(r.modules), func(i int) bool {
			return r.modules[i].Name() >= name
		}); i < len(r.modules) && r.modules[i].Name() == name {
			modules = append(modules, r.modules[i])
		} else {
			return fmt.Errorf("unregistered module %s", name)
		}
	}
	r.modules = modules
	return nil
}

func (r *registry) Initialize(j json.RawMessage) (err error) {
	r.initialized = true

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
					log.Infof("module %s exited with error: %v", m.Name(), err)
					r.Shutdown() // 异常退出的模块将引起服务结束
				}
			} else {
				log.Infof("module %s exited gracefully", m.Name())
			}
		}(m)
	}
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
