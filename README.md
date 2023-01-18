# 如果有一千万个人抢票怎么办？



> 详细介绍请移步：https://bytedancecampus1.feishu.cn/docx/doxcnuQwm7jF76P8o5gU9a0LY1g



### V0
- 确立选型，搭建基本框架

### V1
- 优化了消息队列连接逻辑，初始化连接后复用Channel。不用每次使用都建立新连接
- 瓶颈：并发量加高，服务器拒绝TCP连接，无法发送请求

### V2
- 延时任务取消未支付订单修改为消息队列异步执行，缓解瞬间MySQL压力
- 采用HTTP连接池，解决高并发情况下无法建立连接的情况

### V3 最终版
- 为达到更高的吞吐量，消息队列由RabbitMQ改用Kafka，大大提高有效响应率


- 架构类型：单体架构
- 技术选型
  - 框架：gin gorm
  - 缓存：单机Redis
  - 数据库：单机MySQL
  - 消息队列：Kafka
- 可支持一分钟内同时抢票人数：16万




### 目录结构：
```text
>GrabVotes:                                        
├─client
│      client.go        // 模拟抢票客户端
│
├─cmd
│      main.go      //程序入口
│      table.sql    // 建表语句
│
└─internal
    ├─controller
    │      controller.go        //控制层
    │
    ├─dao
    │  ├─mysql
    │  │      mysql.go      //与MySQL交互
    │  │
    │  └─redis
    │          redis.go     //与Redis交互
    │
    ├─logic
    │      kafka.go     //Kafka消息队列
    │      logic.go     //业务逻辑层
    │      simpleMQ.go  //RabbitMQ消息队列（已弃用）
    │
    ├─model
    │      model.go     //模型层
    │
    └─pkg
        ├─jwt
        │      jwt.go       //用户鉴权中间件
        │
        └─snowid
                snowid.go   //id生成工具
```






