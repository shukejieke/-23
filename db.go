package repo

import (
	"log"
	"sync"

	"gorm.io/gorm"

	"stzbHelper/model"
)

var (
	dbMu       sync.RWMutex
	db         *gorm.DB
	defaultOnce sync.Once
)

// SetDB 注入数据库连接，供 repo 层统一使用。
// 设计目标：后续替换为 DI（依赖注入）时，不需要改所有业务代码。
func SetDB(conn *gorm.DB) {
	dbMu.Lock()
	defer dbMu.Unlock()
	db = conn
}

// DB 获取当前可用连接：
// 1) 优先返回注入连接
// 2) 回退历史全局连接 model.Conn（兼容旧逻辑）
func DB() *gorm.DB {
	dbMu.RLock()
	if db != nil {
		defer dbMu.RUnlock()
		return db
	}
	dbMu.RUnlock()

	if model.Conn != nil {
		return model.Conn
	}

	// 兜底：避免首个请求早于抓包激活时出现 nil panic
	defaultOnce.Do(func() {
		log.Println("[repo] 检测到数据库连接为空，自动初始化默认库 stzb_default")
		model.InitDB("stzb_default")
	})
	return model.Conn
}
