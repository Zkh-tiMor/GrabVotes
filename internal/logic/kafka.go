package logic

import (
	"GrabVotes/internal/dao/mysql"
	"GrabVotes/internal/model"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
)

//Command to start Kafka：
//bin\windows\zookeeper-server-start.bat config\zookeeper.properties
//bin\windows\kafka-server-start.bat config\server.properties
var producerMap map[string]sarama.AsyncProducer
var kafkaClient sarama.Client

//  DO NOT EDIT !!!
const (
	insertOderTopic      string = "insertOrder"
	updateTicketNumTopic string = "updateTicket"
	cancelOrderTopic     string = "cancelOrder"
)

func ProducerClose() {
	producerMap[insertOderTopic].Close()
	producerMap[updateTicketNumTopic].Close()
	producerMap[cancelOrderTopic].Close()
}

func InitKafKa() error {
	config := sarama.NewConfig()
	config.Producer.MaxMessageBytes = 33554432
	var err error
	kafkaClient, err = sarama.NewClient([]string{"localhost:9092"}, config)
	if err != nil {
		return err
	}

	//  初始化生产者，动态类型为 asyncProducer
	producerMap = make(map[string]sarama.AsyncProducer)
	producerMap[insertOderTopic], err = sarama.NewAsyncProducerFromClient(kafkaClient)
	if err != nil {
		return err
	}
	producerMap[updateTicketNumTopic], err = sarama.NewAsyncProducerFromClient(kafkaClient)
	if err != nil {
		return err
	}
	producerMap[cancelOrderTopic], err = sarama.NewAsyncProducerFromClient(kafkaClient)
	if err != nil {
		return err
	}
	// 初始化消费者
	if err = consumeInsertOrder(); err != nil {
		return err
	}
	fmt.Println("Kafka插入订单就绪")
	if err = consumeCancelOrder(); err != nil {
		return err
	}
	fmt.Println("Kafka取消订单就绪")
	if err = consumeTicketNum(); err != nil {
		return err
	}
	fmt.Println("Kafka更新票数就绪")
	return nil
}

func SendMQ(topicName string, msg sarama.Encoder) {
	message := &sarama.ProducerMessage{
		Topic: topicName,
		Value: msg,
	}
	select {
	case producerMap[topicName].Input() <- message:
	case err := <-producerMap[topicName].Errors():
		fmt.Println("kafka发送消息失败：" + err.Error())
	}

}

func consumeInsertOrder() error {
	consumer, err := sarama.NewConsumerFromClient(kafkaClient)
	if err != nil {
		return err
	}
	partitions, err := consumer.Partitions(insertOderTopic)
	if err != nil {
		return err
	}
	for _, partitionID := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(insertOderTopic, partitionID, sarama.OffsetNewest)
		if err != nil {
			return err
		}
		go func(pc *sarama.PartitionConsumer) {
			defer (*pc).Close()
			// block
			for msg := range (*pc).Messages() {
				//  业务逻辑部分
				var dataModel model.OrderModel
				err = json.Unmarshal(msg.Value, &dataModel)
				if err != nil {
					fmt.Println("json反序列化失败：", err)
					continue
				}
				if err = mysql.InsertOrder(dataModel); err != nil {
					fmt.Println("MySQL插入订单失败：", err)
				}
			}
		}(&partitionConsumer)
	}
	return nil
}

func consumeTicketNum() error {
	consumer, err := sarama.NewConsumerFromClient(kafkaClient)
	if err != nil {
		return err
	}
	partitions, err := consumer.Partitions(updateTicketNumTopic)
	if err != nil {
		return err
	}
	for _, partitionID := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(updateTicketNumTopic, partitionID, sarama.OffsetNewest)
		if err != nil {
			return err
		}
		go func(pc *sarama.PartitionConsumer) {
			defer (*pc).Close()
			// block
			for msg := range (*pc).Messages() {
				//  业务逻辑部分
				var dataModel model.MqTicket
				err = json.Unmarshal(msg.Value, &dataModel)
				if err != nil {
					fmt.Println("json反序列化失败：", err)
					continue
				}
				//  调用dao
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
		}(&partitionConsumer)
	}
	return nil
}

func consumeCancelOrder() error {
	consumer, err := sarama.NewConsumerFromClient(kafkaClient)
	if err != nil {
		return err
	}
	partitions, err := consumer.Partitions(cancelOrderTopic)
	if err != nil {
		return err
	}
	for _, partitionID := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(cancelOrderTopic, partitionID, sarama.OffsetNewest)
		if err != nil {
			return err
		}
		go func(pc *sarama.PartitionConsumer) {
			defer (*pc).Close()
			// block
			for msg := range (*pc).Messages() {
				//  业务逻辑部分
				var cancel model.CancelOrderMq
				err := json.Unmarshal(msg.Value, &cancel)
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
		}(&partitionConsumer)
	}
	return nil
}
