<div align="center">

# 极简版抖音

[![Visitors](https://api.visitorbadge.io/api/daily?path=https://github.com/chuxin0816/douyin&label=VISITORS%20TODAY&countColor=%231758f0)](https://github.com/chuxin0816/douyin)
</div>

该项目为字节跳动青训营项目，实现了抖音主要功能模块，包括视频流、视频投稿、注册登录、个人信息、点赞评论、用户关注、即时通讯等核心业务，从[单体架构](https://github.com/chuxin0816/douyin/tree/v1)升级到[微服务架构](https://github.com/chuxin0816/douyin)，并持续优化相关技术实现。

接口文档: [Apifox](https://apifox.com/apidoc/shared-0c80e0c6-daca-4b12-96a4-01ca8c2b6cd1) ｜ 项目演示地址：http://chuxin0816.com:8888/ (已关闭)
## 项目部署
`docker-compose up -d`
## 项目结构：
```shell
. #篇幅有限，只展示部分目录
├── api                 HTTP服务
├── cmd
│   ├── docker     
│   └── gen             Gorm/Gen
├── config         
├── dal                 访问数据库代码(MySQL, MongoDB, Redis)
├── logger         
├── pkg
│   ├── jwt
│   ├── kafka
│   ├── oss
│   ├── snowflake
│   └── tracing
├── rpc
│   ├── client          RPC客户端
│   ├── idl        
│   ├── kitex_gen
│   └── service         RPC服务端
├── docker-compose.yml    
├── douyin.sql          MySQL表结构
└── message.js          MongoDB表结构
```
> 请求链路: http请求->api/router->api/controller->rpc/client->rpc/service->dal
##  性能测试
使用wrk进行性能测试，400个连接，16个线程，压力测试30s：读接口QPS 3500+，写接口QPS 2800+，压测过程全链路无错误
## 技术选型：
- 框架选型：使用 **Hertz** 作为 HTTP 微服务框架，**Kitex** 作为 RPC 微服务框架；使用 **GORM GEN** 生成代码并操作 MySQL 数据库，具备简单易用、防SQL注入等特性。
- 数据库：使用 **MySQL** 和 **MongoDB** 存储数据，使用 **Redis** 作为分布式缓存。
- 服务注册与发现：使用 **Consul** 作为服务发现与注册中心和配置中心，并通过 **viper** 实时监控和重新读取配置文件。
- 缓存策略：通过**旁路缓存**和**写回**策略提升数据访问速度并减轻数据库压力。使用 **SingleFlight** 减轻 Redis 压力和防止缓存击穿、使用布隆过滤器减少缓存穿透，并通过随机延时策略避免缓存雪崩。
- 中间件：采用**令牌桶**作为限流中间件，**JWT** 作为用户认证中间件，使用 **Kafka** 作为消息队列，实现对高频的点赞和评论操作削峰、配合 **Canal** 删除缓存、同步布隆过滤器等操作
- 云原生：通过 **OpenTelemetry** + **Jaeger** 实现分布式链路追踪，使用 **Docker Compose** 快速部署项目，并通过 **GitHub Actions** 自动构建和推送镜像
- 其他：使用 **Snowflake** 算法生成全局唯一ID，使用 **ffmpeg** 截取视频第5帧作为封面，使用 **OSS** 存储视频和视频封面
## 代码生成示例:
```shell
1. Gorm/Gen代码生成
go run cmd/gen/generator.go
2. Kitex代码生成
cd rpc/service/feed
kitex -module douyin -service feed -gen-path ../../kitex_gen/ ../../idl/feed.thrift
```
## 未来更新:
* 使用ElasticSearch对用户消息和系统日志进行索引存储
* 使用Prometheus和Grafana监控服务
* 使用Kubernetes编排容器