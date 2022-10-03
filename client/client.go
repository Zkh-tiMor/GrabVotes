package client

import (
	"crypto/tls"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mySecret = []byte("GrabVotes")
var cli = &http.Client{}
var cnt float32

type myClaims struct {
	UserID string `json:"id"`
	jwt.StandardClaims
}

func main() {
	initClient()
	cnt = 0
	wg := &sync.WaitGroup{}
	wg.Add(4)
	beginTime := time.Now()
	go doGrab(0, 40000, wg)
	go doGrab(40001, 80000, wg)
	go doGrab(80001, 120000, wg)
	go doGrab(120001, 160000, wg)
	wg.Wait()
	endTime := time.Now()
	fmt.Printf("请求失败率为：%f%% 耗时：%f秒\n 失败次数：%d", cnt*100/float32(160000), endTime.Sub(beginTime).Seconds(), int(cnt))
}

func doGrab(begin int, end int, group *sync.WaitGroup) {
	defer group.Done()
	//  同步组
	g := &sync.WaitGroup{}
	g.Add(end - begin + 1)
	//  并发请求
	for i := begin; i <= end; i++ {
		token, err := generateToken(strconv.Itoa(i))
		if err != nil {
			panic("生成token失败：" + err.Error())
		}
		time.Sleep(time.Nanosecond)
		go func(*sync.WaitGroup, string) {
			err := sendRequest(g, token)
			if err != nil {
				l := &sync.Mutex{}
				l.Lock()
				cnt++
				l.Unlock()
				fmt.Println("请求出错：", err.Error())
			}
		}(g, token)
	}
	g.Wait()
}

func sendRequest(wg *sync.WaitGroup, token string) error {
	// 1.配置请求参数
	grabUrl := url.Values{}
	grabUrl.Set("token", token)
	grabUrl.Set("ticketID", "zkh_mirror")
	//  2.捏造post请求
	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/GrabAction", strings.NewReader(grabUrl.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(grabUrl.Encode())))
	//  3.发送请求
	resp, err := cli.Do(req)
	defer wg.Done()
	if err != nil {
		return err
	}
	ioutil.ReadAll(resp.Body) //如果不读取数据连接无法复用！！
	return nil
}

// 初始化连接
func initClient() {
	//  连接复用需要server端和client同时支持
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		//InsecureSkipVerify用来控制客户端是否证书和服务器主机名。如果设置为true, 则不会校验证书以及证书中的主机名和服务器主机名是否一致。
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, //拨号等待连接完成的最大时间
			KeepAlive: 30 * time.Second, //保持网络连接活跃keep-alive探测间隔时间。
		}).DialContext,
		MaxIdleConns:        5000,
		MaxConnsPerHost:     5000,
		MaxIdleConnsPerHost: 5000,
		IdleConnTimeout:     300 * time.Second,
	}
	cli = &http.Client{
		Transport: tr,
		Timeout:   9 * time.Second, //设置超时，包含connection时间、任意重定向时间、读取response body时间
	}
}

func generateToken(userID string) (Token string, err error) {
	// 建立自己的token字段
	c := myClaims{
		userID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}
	// 加密并获得的完整编码后的aToken
	if Token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret); err != nil {
		return "", err
	}
	return
}
