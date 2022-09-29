package main

import (
	"GrabVotes/internal/controller"
	"GrabVotes/internal/dao/mysql"
	"GrabVotes/internal/dao/redis"
	"GrabVotes/internal/logic"
	"GrabVotes/internal/pkg/snowid"
	"github.com/gin-gonic/gin"
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
	if err := logic.SetRedisTicketNum("zkh_mirror", 2000); err != nil {
		panic("初始化Redis缓存失败：" + err.Error())
	}

	// 初始化消息队列
	if err := logic.InitMqQueue([]string{logic.UpdateTicketNum, logic.InsertMysqlOrder}); err != nil {
		panic("初始化消息队列失败：" + err.Error())
	}
}

func main() {
	r := gin.Default()

	gin.SetMode(gin.DebugMode)
	// 抢购订单，感觉需要加消息队列削峰了
	r.POST("/GrabAction", controller.JWTAuth, controller.GrabAction)
	// 支付订单
	r.POST("/defrayAction", controller.JWTAuth, controller.DefrayAction)

	go func() {
		logic.GetConsumer().UpdateTicketNum(logic.UpdateTicketNum)
	}()

	go func() {
		logic.GetConsumer().InsertMysqlOrder(logic.InsertMysqlOrder)
	}()
	err := r.Run(":8080")
	if err != nil {
		panic("启动异常：" + err.Error())
	}
}
