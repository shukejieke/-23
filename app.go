package app

import (
	"log"
	"sync"
	"time"

	"github.com/google/gopacket/pcap"

	"stzbHelper/internal/capture"
	"stzbHelper/internal/repo"
	"stzbHelper/internal/server/httpserver"
	"stzbHelper/internal/service"
	"stzbHelper/model"
)

// Run 是应用启动入口（重构后的总编排层）。
// 职责：
// 1) 探测网卡
// 2) 启动 HTTP 服务
// 3) 启动额外数据导入流程
// 4) 为每张网卡启动抓包协程
func Run() {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal("无法获取网络接口列表:", err)
	}
	if len(devices) == 0 {
		log.Fatal("未找到可用的网络接口")
	}

	var wg sync.WaitGroup

	// 先初始化默认库，避免首个激活包未到达前写库失败
	model.InitDB("stzb_default")
	repo.SetDB(model.Conn)

	// 启动 JSON -> MySQL 的导入逻辑，完成后回填历史战报名称
	go func() {
		model.ImportExtraJSONToMySQL()
		service.BackfillBattleReportNames()
	}()

	// 启动 HTTP 控制面
	go httpserver.Start(&wg)
	wg.Add(1)

	log.Println("stzbHelper开始运行!")
	log.Println("等待打开主公簿激活软件...")
	log.Println("未打开主公簿激活软件前软件可能会出现报错！")
	time.Sleep(100 * time.Millisecond)

	// 每个网卡开启一个抓包协程，避免单网卡瓶颈
	for _, device := range devices {
		wg.Add(1)
		go capture.Start(device.Name, &wg)
	}

	wg.Wait()
}
