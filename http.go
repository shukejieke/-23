package httpserver

import (
	"io"
	"log"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	httpRoute "stzbHelper/http"
)

// Start 启动 HTTP 服务（控制面）。
// 说明：
// - 当前统一监听 9527
// - 路由注册仍复用原 http.RegisterRoute
func Start(wait *sync.WaitGroup) {
	log.Println("HTTP服务启动")
	gin.DefaultWriter = io.Discard
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())

	httpRoute.RegisterRoute(r)

	log.Println("http://127.0.0.1:9527 浏览器打开此地址控制软件")

	if err := r.Run(":9527"); err != nil {
		log.Fatal("http服务启动失败:" + err.Error())
		wait.Done()
		return
	}
}
