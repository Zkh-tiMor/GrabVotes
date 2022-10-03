package logic

import (
	"GrabVotes/internal/dao/redis"
	"GrabVotes/internal/model"
	"GrabVotes/internal/pkg/snowid"
	"encoding/json"
	"github.com/Shopify/sarama"
	"gorm.io/gorm/utils"
	"time"
)

const (
	payTime      = time.Minute * 20
	incr    int8 = 1
	decr    int8 = -1
)

// GrabAction 抢票成功返回true,nil
func GrabAction(userID string, ticketID string) (bool, error) {
	// 校验ticketID的正确性
	ticketExist, err := redis.Dealer().ExistKey(ticketID)
	if err != nil {
		return false, err
	}
	if !ticketExist { //票不存在
		return false, nil
	}
	//  DecrBy 返回减一后剩余的票数，保证原子性
	remain, err := redis.Dealer().DecrByTicket(ticketID, 1)
	if err != nil {
		panic("Redis DecrBy failed：" + err.Error())
	}
	if remain < 0 { //票已经卖完了
		return false, nil
	}
	//  到这就是抢票成功了，还要操作数据库票数减1，用消息队列异步更新数据库
	order := model.OrderModel{
		OrderID:  snowid.GenID(),
		TicketID: ticketID,
		UserID:   userID,
	}
	mqTicket := model.MqTicket{
		TicketID: utils.ToString(ticketID),
		Style:    decr,
		Amount:   1,
	}
	cancelOrderMq := model.CancelOrderMq{
		OrderID: order.OrderID,
	}
	//  序列化
	mqTicketByte, err := json.Marshal(mqTicket)
	if err != nil {
		return false, err
	}
	orderByte, err := json.Marshal(order)
	if err != nil {
		return false, err
	}
	cancelOrderByte, err := json.Marshal(cancelOrderMq)
	if err != nil {
		return false, err
	}
	//  异步MySQL插入订单
	SendMQ(insertOderTopic, sarama.ByteEncoder(orderByte))
	//  异步更新票数
	SendMQ(updateTicketNumTopic, sarama.ByteEncoder(mqTicketByte))

	//  开启定时任务：若20分钟后未支付则取消订单
	go func(cancelOrderModel []byte) { //消除闭包影响
		timer := time.NewTimer(payTime)
		_ = <-timer.C //阻塞20分钟后，修改订单
		SendMQ(cancelOrderTopic, sarama.ByteEncoder(cancelOrderModel))
	}(cancelOrderByte)
	return true, nil
}

func SetRedisTicketNum(TicketID string, num int) error {
	return redis.Dealer().SetTicketNum(TicketID, num)
}
