package mysql

import (
	"GrabVotes/internal/dao/redis"
	"GrabVotes/internal/model"
	"GrabVotes/internal/pkg/conf"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	username     string = "root"
	password     string = conf.MySQLPassword
	host         string = "127.0.0.1"
	port         string = "3306"
	databaseName string = "grab"

	orderPaid   int8 = 1
	orderUnpaid int8 = 0
)

var db *gorm.DB

func InitMysql() (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, databaseName,
	)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return err
	}
	sqldb, _ := db.DB()
	if err = sqldb.Ping(); err != nil {
		fmt.Println("MySQL ping failed...")
	}
	return
}

func InsertOrder(model model.OrderModel) error {
	err := db.Create(model).Error
	if err != nil {
		return err
	}
	return nil
}

// DeleteOrder 取消订单
func DeleteOrder(orderID string) error {
	//  先以update形式查询订单status，如果status为1，就直接返回，status为0，就修改
	return db.Transaction(func(tx *gorm.DB) error {
		//  锁获取了之后只有在事务结束才释放，select添加写锁防止事务冲突
		//  查询判断订单状态
		order := model.OrderModel{}
		if err := tx.Raw("select `status`,`ticket_id` from `order` where `order_id` = ? for update ", orderID).Scan(&order).Error; err != nil {
			return err
		}
		if order.Status == orderPaid { //订单已支付直接返回
			return nil
		}
		//  仍未支付，软删除订单
		if err := tx.Exec("update `order` set `deleted` = 1 where `order_id` = ? and `status` = 0", orderID).Error; err != nil {
			return err
		}
		//  MySQL中订单数量加1
		if err := tx.Exec("update ticket_msg set `ticket_num` = `ticket_num`+1 where `ticket_id` = ?", order.TicketID).Error; err != nil {
			return err
		}
		// Redis中订单数量加1
		_, err := redis.Dealer().IncrByTicket(redis.TicketNumPre, 1)
		if err != nil {
			return err
		}
		return nil
	})

}

func IncrTicketNum(ticketID string, num int) error {
	if err := db.Exec("update ticket_msg set `ticket_num` = `ticket_num` + ? where `ticket_id` = ?", num, ticketID).Error; err != nil {
		return err
	}
	return nil
}

func DecrTicketNum(ticketID string, num int) error {
	if err := db.Exec("update `ticket_msg` set `ticket_num` = `ticket_num` - ? where `ticket_id` = ?", num, ticketID).Error; err != nil {
		return err
	}
	return nil
}
