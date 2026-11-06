package main

import (
	"github.com/David89Lee/pyrolysis_ty_server/common"
	"github.com/David89Lee/pyrolysis_ty_server/middleware"
	"github.com/David89Lee/pyrolysis_ty_server/model"
	"github.com/David89Lee/pyrolysis_ty_server/router"
	"go.uber.org/zap"

	"os"
	"os/signal"
	"syscall"
)

func init() {
	common.SetUpSetting("conf")
	middleware.InitLogger()
	model.Setup()
}

// @title 台谊热解粒子测试
// @version 1.0
// @description 台谊热解粒子测试
// @contact.name LG
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host 127.0.0.1:8081

func main() {

	defer func(l *zap.Logger) {
		err := l.Sync()
		if err != nil {
			zap.L().Error("zap日志同步失败", zap.Error(err))
		}
	}(zap.L())

	// http服务
	r := router.InitRouter()
	// start server in a goroutine
	go func() {
		if err := r.Run(common.ServerSetting.HttpPort); err != nil {
			zap.L().Error("启动服务失败", zap.Error(err))
		}
	}()

	// DTU 通过 TCP 主动上报数据，独立运行，失败时只记日志，不影响 HTTP / SSE 服务
	go startTcpServer(common.ServerSetting.TcpPort)

	// 每天凌晨 5 点抽稀历史数据
	startHistoryJob()

	// 阻塞主线程，等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	zap.L().Info("收到退出信号，服务停止")
}
