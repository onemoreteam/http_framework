package log

import (
	"encoding/json"

	"github.com/ntons/log-go/config"
	"github.com/onemoreteam/httpframework/modularity"
)

func init() {
	modularity.Register(&logModule{})
}

type logModule struct {
	modularity.Skeleton
}

func (logModule) Name() string { return "log" }

// 必须最先初始化
func (logModule) Priority() int { return -100 }

func (m logModule) Initialize(jb json.RawMessage) (err error) {
	if jb != nil {
		var cfg config.Config
		if err = json.Unmarshal(jb, &cfg); err != nil {
			return
		}
		if err = cfg.Use(); err != nil {
			return
		}
	}
	return
}
