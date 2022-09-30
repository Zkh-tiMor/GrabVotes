package main

import (
	"GrabVotes/internal/controller"
	"GrabVotes/internal/dao/mysql"
	"GrabVotes/internal/dao/redis"
	"GrabVotes/internal/logic"
	"GrabVotes/internal/pkg/snowid"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func init() {
	if err := mysql.InitMysql(); err != nil {
		panic("MySQL初始化失败：" + err.Error())
	}

	if err := redis.InitRedis(); err != nil {
		panic("Redis初始化失败：" + err.Error())
	}

	if err := snowid.Init(); err != nil {
		panic("snowflake初始化失败：" + err.Error())
	}

	// 初始化Redis缓存中的票数
	if err := logic.SetRedisTicketNum("zkh_mirror", 5000); err != nil {
		panic("初始化Redis缓存失败：" + err.Error())
	}

	// 初始化消息队列
	if err := logic.InitMqQueue([]string{logic.UpdateTicketNum, logic.InsertMysqlOrder, logic.CancelOrder}); err != nil {
		panic("初始化消息队列失败：" + err.Error())
	}
}

func main() {
	r := gin.Default()

	gin.SetMode(gin.DebugMode)
	// 抢购订单
	r.POST("/GrabAction", controller.JWTAuth, controller.GrabAction)
	// 支付订单
	r.POST("/defrayAction", controller.JWTAuth, controller.DefrayAction)

	go func() {
		logic.GetConsumer().UpdateTicketNum(logic.UpdateTicketNum)
	}()

	go func() {
		logic.GetConsumer().InsertMysqlOrder(logic.InsertMysqlOrder)
	}()

	go func() {
		logic.GetConsumer().CancelOrder(logic.CancelOrder)
	}()

	server := http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic("启动异常：" + err.Error())
	}
}
