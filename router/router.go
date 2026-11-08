package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	_ "github.com/David89Lee/pyrolysis_ty_server/docs"
	"github.com/David89Lee/pyrolysis_ty_server/middleware"
	v1 "github.com/David89Lee/pyrolysis_ty_server/router/api/v1"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())

	// 使用zap记录日志
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		zap.L().Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
		)
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	apiv1 := r.Group("/api/v1")

	apiv1.POST("/auth/login", v1.Login)

	// 以下接口需登录后携带 token 访问
	authed := apiv1.Group("", middleware.AuthRequired())

	authed.POST("/info", v1.CreateInfo)
	authed.PUT("/info", v1.UpdateInfo)
	authed.DELETE("/info/:id", v1.DeleteInfo)
	authed.GET("/info", v1.FindInfo)

	// 采样数据实时推送（SSE），token 通过查询参数传递
	authed.GET("/sample/stream", v1.SampleStream)

	// 查询采样数据：dtu 为 1~4，records 实时表，history 历史表，curve 单参数曲线
	authed.GET("/dtu/:dtu/records", v1.FindDtuRecords)
	authed.GET("/dtu/:dtu/history", v1.FindDtuHistory)
	authed.GET("/dtu/:dtu/curve", v1.FindDtuCurve)

	return r
}
