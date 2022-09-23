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

const MqName = "updateTicketNum"

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	QueueName string //队列名称
	Exchange  string //交换机名称
	Key       string //bind Key 名称
	MqUrl     string //连接信息
}

func (r *RabbitMQ) CloseMq() {
	r.conn.Close()
	r.channel.Close()
}

// NewRabbitMQSimple 创建简单模式下RabbitMQ实例
func NewRabbitMQSimple(queueName string) (*RabbitMQ, error) {
	//创建RabbitMQ实例
	rabbitmq := &RabbitMQ{
		QueueName: queueName,
		Exchange:  "",
		Key:       "",
		MqUrl:     mqUrl,
	}
	var err error
	//获取connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.MqUrl)
	if err != nil {
		fmt.Println("获取connection失败：", err.Error())
		return &RabbitMQ{}, err
	}
	//获取channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	if err != nil {
		fmt.Println("获取channel失败：", err.Error())
		return &RabbitMQ{}, err
	}
	return rabbitmq, nil
}

func (r *RabbitMQ) PublishSimple(ModelJson []byte) error {
	//1.申请队列，如果队列不存在会自动创建，存在则跳过创建
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		false, //是否持久化
		false, //是否自动删除
		false, //是否具有排他性
		false, //是否阻塞处理
		nil,   //额外的属性
	)
	if err != nil {
		fmt.Println("申请队列失败：", err.Error())
		return err
	}
	//调用channel 发送消息到队列中
	err = r.channel.Publish(
		r.Exchange,
		r.QueueName,
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

func (r *RabbitMQ) ConsumeSimple() {
	//1.申请队列，如果队列不存在会自动创建，存在则跳过创建
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		false, //是否持久化
		false, //是否自动删除
		false, //是否具有排他性
		false, //是否阻塞处理
		nil,   //额外的属性
	)
	if err != nil {
		panic("申请队列失败：" + err.Error())
	}
	//接收消息
	msgs, err := r.channel.Consume(
		q.Name,
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

	fmt.Println("消息队列消费端已就绪")
	for d := range msgs { // 如果没有消息会一直阻塞，直到channel关闭

		var dataModel model.MqModel
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
