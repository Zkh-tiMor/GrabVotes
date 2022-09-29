# What would happen if 10 million people grabbing tickets?

V0.0.0

- 粗略的尝试，简单测试后只能抗住1e3的量


V0.0.1    量级：5000

- 优化了消息队列连接逻辑，初始化连接后复用Channel。不用每次使用都建立新连接
- 目前问题是并发量加高，服务器拒绝TCP连接，那么也就无法建立HTTP连接了。瓶颈在客户端跟服务器的连接上 （离及格还差了十万八千里