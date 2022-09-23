package redis

import (
	"GrabVotes/internal/pkg/conf"
	"fmt"
	"github.com/go-redis/redis"
)

var redisDB *redis.Client

const (
	TicketNumPre string = "ticketNumbers"
)

func InitRedis() (err error) {
	redisDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",   // redis地址
		Password: conf.RedisPassword, // redis密码，没有则留空
		DB:       0,                  // 默认数据库，默认是0
	})

	//通过 *redis.Client.Ping() 来检查是否成功连接到了redis服务器
	_, err = redisDB.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

type RedisFunc interface {
	ExistKey(ticketID string) (bool, error)
	SetTicketNum(ticketID string, votes int) error
	IncrByTicket(ticketID string, incr int64) (int64, error)
	DecrByTicket(ticketID string, decr int64) (int64, error)
}

type redisDealer struct{}

func (r *redisDealer) DecrByTicket(ticketID string, decr int64) (int64, error) {
	num, err := redisDB.DecrBy(fmt.Sprintf("%s:%s", TicketNumPre, ticketID), decr).Result()
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (r *redisDealer) IncrByTicket(ticketID string, incr int64) (int64, error) {
	num, err := redisDB.IncrBy(fmt.Sprintf("%s:%s", TicketNumPre, ticketID), incr).Result()
	if err != nil {
		return 0, err
	}
	return num, nil
}

var _ RedisFunc = &redisDealer{}

func Dealer() *redisDealer {
	return &redisDealer{}
}

func (r *redisDealer) SetTicketNum(ticketID string, ticketNUM int) error {
	if err := redisDB.Set(fmt.Sprintf("%s:%s", TicketNumPre, ticketID), ticketNUM, 0).Err(); err != nil {
		return err
	}
	return nil
}

// ExistKey 检查key是否存在，存在返回true
func (r *redisDealer) ExistKey(ticketID string) (bool, error) {
	exist := int64(1)
	keyStatus, err := redisDB.Exists(fmt.Sprintf("%s:%s", TicketNumPre, ticketID)).Result()
	if err != nil {
		return false, err
	}
	return keyStatus == exist, nil
}
