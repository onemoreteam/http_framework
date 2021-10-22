package modularity

import (
	"encoding/json"
)

/// 逻辑模块
type Module interface {
	// 模块名称
	Name() string
	// 优先级
	Priority() int
	/// 初始化
	Initialize(json.RawMessage) (err error)
	/// 清理资源
	Finalize()
	/// 启动服务
	Serve() (err error)
	/// 关闭模块，方法应该在标记停止后立即返回
	Shutdown()
}
