package logic

import (
	"GrabVotes/internal/dao/mysql"
	"GrabVotes/internal/dao/redis"
	"GrabVotes/internal/model"
	"GrabVotes/internal/pkg/snowid"
	"encoding/json"
	"gorm.io/gorm/utils"
	"log"
	"time"
)

//  Redis存的数据MySQL一定也要有

//  1. 用户发起抢购
//  2.抢购失败直接返回失败，抢票成功则生成订单（20分钟内未支付则取消订单，票数加1）
//  订单写入数据库并开启延时任务，20分钟后如果status字段为0（未支付）则软删除，deleted位 --> 置1

//  只有支付成功了，ticket_msg中票的数量才会减1

//  策略：先更新MySQL，再更新缓存（会有一点不一致，以MySQL为准）
//  3.用户支付订单后写入数据库

const (
	payTime      = time.Minute * 20
	incr    int8 = 1
	decr    int8 = -1
)

// GrabAction 抢票成功返回true,nil
func GrabAction(userID string, ticketID string) (bool, error) {
	// 先校验ticketID的正确性
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
	//  到这就是抢票成功了，还要操作数据库票数减1，用消息队列异步更新数据库，否则导致锁争用
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
	mqTicketByte, err := json.Marshal(mqTicket)
	if err != nil {
		return false, err
	}
	orderByte, err := json.Marshal(order)
	if err != nil {
		return false, err
	}
	//  MySQL插入订单
	//  TODO:对这里来说，Redis并没有为MySQL挡住流量
	if err = GetPublisher().JsonByte(orderByte, InsertMysqlOrder); err != nil {
		return false, err
	}
	//  异步消息队列
	if err = GetPublisher().JsonByte(mqTicketByte, UpdateTicketNum); err != nil {
		return false, err
	}

	//  插入成功，开启定时任务：若20分钟后未支付则取消订单
	go func(orderID string) { //消除闭包影响
		timer := time.NewTimer(payTime)
		beginTime := <-timer.C //阻塞20分钟后，修改订单
		if err := mysql.DeleteOrder(orderID); err != nil {
			log.Println("取消订单失败：", err.Error(), "time:", beginTime)
		}
	}(order.OrderID)
	return true, nil
}

func SetRedisTicketNum(TicketID string, num int) error {
	return redis.Dealer().SetTicketNum(TicketID, num)
}
