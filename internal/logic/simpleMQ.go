package logic

import (
	"GrabVotes/internal/dao/mysql"
	"GrabVotes/internal/model"
	"GrabVotes/internal/pkg/conf"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

const (
	mqUsername    string = "zhangkh"
	mqPassword    string = conf.RabbitMQPassword
	mqVirtualHost string = "grab"
)

// TODO: 瓶颈看起来像在amqp的连接上， 把每次队列操作的QueueDeclare操作独立出来
var mqUrl string = fmt.Sprintf("amqp://%s:%s@127.0.0.1:5672/%s", mqUsername, mqPassword, mqVirtualHost)

const UpdateTicketNum = "updateTicketNum"
const InsertMysqlOrder = "insertMysqlOrder"
const CancelOrder = "cancelOrder"

var rabbitMqQueue *amqp.Channel

func InitMqQueue(queueNames []string) error {
	conn, err := amqp.Dial(mqUrl)
	if err != nil {
		return err
	}
	rabbitMqQueue, err = conn.Channel() //初始化channel
	if err != nil {
		return err
	}

	for _, queueName := range queueNames {
		_, err = rabbitMqQueue.QueueDeclare( //声明队列
			queueName,
			true,  //是否持久化
			false, //是否自动删除
			false, //是否具有排他性
			false, //是否阻塞处理
			nil,   //额外的属性
		)
		if err != nil {
			return err
		}
	}

	return nil
}

type publisher struct{}

func GetPublisher() *publisher {
	return &publisher{}
}

func (p *publisher) JsonByte(ModelJson []byte, queueName string) error {
	//调用channel 发送消息到队列中
	err := rabbitMqQueue.Publish(
		"",
		queueName,
		false, //如果为true，根据自身exchange类型和routekey规则无法找到符合条件的队列会把消息返还给发送者
		false, //如果为true，当exchange发送消息到队列后发现队列上没有消费者，则会把消息返还给发送者
		amqp.Publishing{
			ContentType: "application/json",
			Body:        ModelJson,
		})
	if err != nil {
		fmt.Println("RabbitMQ发送消息失败：", err.Error())
		return err
	}
	return nil
}

type consumer struct{}

func GetConsumer() *consumer {
	return &consumer{}
}

func (c *consumer) InsertMysqlOrder(queueName string) {
	msgs, err := rabbitMqQueue.Consume(
		queueName,
		"",    // consumer, 用来区分多个消费者
		true,  // auto-ack,是否自动应答
		false, // exclusive,是否独有
		false, // no-local，true表示不能将同一个Connection中生产者发送的消息传递给这个Connection中的消费者
		false, // no-wait, 列是否阻塞
		nil,   // args
	)
	if err != nil {
		panic("RabbitMQ接收消息失败：" + err.Error())
	}

	fmt.Println("消息队列插入订单已就绪")
	for d := range msgs { // 如果没有消息会一直阻塞，直到channel关闭

		var dataModel model.OrderModel
		err = json.Unmarshal(d.Body, &dataModel)
		if err != nil {
			fmt.Println("json反序列化失败：", err)
			continue
		}
		if err = mysql.InsertOrder(dataModel); err != nil {
			fmt.Println("MySQL插入订单失败：", err)
		}
	}
}

func (c *consumer) UpdateTicketNum(queueName string) {
	//接收消息
	msgs, err := rabbitMqQueue.Consume(
		queueName,
		"",    // consumer, 用来区分多个消费者
		true,  // auto-ack,是否自动应答
		false, // exclusive,是否独有
		false, // no-local，true表示不能将同一个Connection中生产者发送的消息传递给这个Connection中的消费者
		false, // no-wait, 列是否阻塞
		nil,   // args
	)
	if err != nil {
		panic("RabbitMQ接收消息失败：" + err.Error())
	}

	fmt.Println("消息队列更新票数已就绪")
	for d := range msgs { // 如果没有消息会一直阻塞，直到channel关闭

		var dataModel model.MqTicket
		err = json.Unmarshal(d.Body, &dataModel)
		if err != nil {
			fmt.Println("json反序列化失败：", err)
			continue
		}
		//  直接操作MySQL层
		if dataModel.Style == incr {
			err = mysql.IncrTicketNum(dataModel.TicketID, 1)
			if err != nil {
				log.Println("MySQL票数加1失败", err)
			}
		} else if dataModel.Style == decr {
			err = mysql.DecrTicketNum(dataModel.TicketID, 1)
			if err != nil {
				log.Println("MySQL票数减1失败", err)
			}
		}

	}
}

func (c *consumer) CancelOrder(queueName string) {
	msgs, err := rabbitMqQueue.Consume(
		queueName,
		"",    // consumer, 用来区分多个消费者
		true,  // auto-ack,是否自动应答
		false, // exclusive,是否独有
		false, // no-local，true表示不能将同一个Connection中生产者发送的消息传递给这个Connection中的消费者
		false, // no-wait, 列是否阻塞
		nil,   // args
	)
	if err != nil {
		panic("RabbitMQ接收消息失败：" + err.Error())
	}
	fmt.Println("异步取消订单队列已就绪")

	for d := range msgs { // 如果没有消息会一直阻塞，直到channel关闭
		var cancel model.CancelOrderMq
		err := json.Unmarshal(d.Body, &cancel)
		if err != nil {
			log.Println("反序列化失败：", err.Error())
			continue
		}
		err = mysql.DeleteOrder(cancel.OrderID)
		if err != nil {
			log.Println("取消订单失败：", err.Error())
			continue
		}
	}
}
