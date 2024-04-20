# 抖音
> 字节跳动青训营项目，简易版抖音，实现了抖音的基本功能
## 接口文档：https://apifox.com/apidoc/shared-0c80e0c6-daca-4b12-96a4-01ca8c2b6cd1
## 项目演示地址：http://chuxin0816.com:8888/ (已关闭)
### 项目部署
项目使用docker-compose部署，在配置好config/config.yaml后运行`cd && docker-compose up -d`即可启动项目
> 如果内存不足可以分批构建后启动:
```shell
cd cmd/docker && 
docker-compose build api &&
docker-compose build user &&
docker-compose build favorite &&
docker-compose build comment &&
docker-compose build publish &&
docker-compose build relation &&
docker-compose build message &&
docker-compose up -d
```
初次启动部分服务会因为MySQL没有相关表而报错，需通过douyin.sql建表后再次启动失败的服务
******
## 项目结构：
http请求->api/router->api/controller->rpc/client->rpc/service->dal
##  性能测试
> 使用wrk进行性能测试，400个连接，16个线程，压力测试20s：读接口平均QPS 3000+，写接口平均QPS 2000+
## 技术选型：
* 使用Hertz作为http微服务框架，具有高性能，高可用，高扩展性的特点
* 使用Kitex作为rpc微服务框架，具有高性能、强可扩展的特点
* 使用consul作为服务发现和注册中心
* 使用GORM GEN操作Mysql，具有简单易用，防SQL注入等优点
* 使用Redis作为缓存，提高访问速度，使用定时同步缓存保证数据一致性
* 使用canal订阅Mysql的binlog，发送到kafka异步删除点赞和关注关系缓存
* 使用kafka作为消息队列，对于高频的点赞和评论异步写入数据库，对于点赞数，粉丝数等数量缓存定时同步到数据库
* 使用布隆过滤器防止缓存穿透，使用随机延时防止缓存雪崩
* 使用SingleFlight防止缓存击穿
* 使用Kitex集成的zap日志库记录日志
* 使用JWT作为用户认证，使用中间件进行认证
* 使用bcrypt加密用户密码
* 使用snowflake生成各种id，使用uuid生成oss文件名
* 使用阿里云oss存储视频文件
* 使用ffmpeg进行视频转码和生成封面
* 使用令牌桶作为限流中间件
* 使用viper读取配置文件
* 使用OpenTelemetry+Jaeger进行分布式链路追踪
* 使用docker-compose部署项目
## 代码生成:
```shell
cd rpc/service/feed
kitex -module douyin -service feed -gen-path ../../kitex_gen/ ../../idl/feed.thrift
```
## 未来更新:
* 使用ElasticSearch对用户消息和系统日志进行索引存储
* 使用Prometheus和Grafana监控服务
* 使用Kubernetes编排容器