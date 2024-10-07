<div align="center">

# 极简版抖音

[![Visitors](https://api.visitorbadge.io/api/daily?path=https://github.com/chuxin0816/douyin&label=VISITORS%20TODAY&countColor=%231758f0)](https://github.com/chuxin0816/douyin)
</div>

该项目为字节跳动青训营项目，实现了抖音主要功能模块，包括视频、用户、点赞、评论、关注、消息等服务，从[单体架构](https://github.com/chuxin0816/douyin/tree/v1)升级到[微服务架构](https://github.com/chuxin0816/douyin)，并持续优化相关技术实现。

接口文档: [Apifox](https://apifox.com/apidoc/shared-0c80e0c6-daca-4b12-96a4-01ca8c2b6cd1) ｜ 项目演示地址：http://chuxin0816.com:8888/ (已关闭)
## 项目部署
`docker-compose up -d`
## 项目结构：
```shell
. #篇幅有限，只展示主要部分
├── cmd
│   ├── docker              Dockerfile
│   └── gen                 Gorm/Gen
├── deploy                  Kubernetes部署文件
├── etc                     配置文件
├── src
│   ├── client              RPC客户端
│   ├── common             
│   │   ├── clientsuite     客户端套件
│   │   ├── jwt             JWT认证
│   │   ├── kafka           Kafka消息队列
│   │   ├── mtl             指标监控、链路追踪、日志
│   │   ├── oss             阿里云OSS
│   │   ├── serversuite     服务端套件
│   │   ├── snowflake       雪花算法
│   ├── config
│   ├── dal                 访问数据库(PostgreSQL, Redis, NebulaGraph)
│   ├── idl
│   ├── kitex_gen
│   └── service
│       ├── api             API服务
│       ├── comment         评论服务
│       ├── favorite        点赞服务 
│       ├── message         消息服务
│       ├── relation        关注服务
│       ├── user            用户服务
│       └── video           视频服务
└── docker-compose.yml
```
##  性能测试
使用wrk进行性能测试，400个连接，16个线程，压力测试30s：读接口QPS 3500+，写接口QPS 2800+，压测过程全链路无错误
## 技术选型：
- 框架选型：使用 **Hertz** 作为 HTTP 微服务框架，**Kitex** 作为 RPC 微服务框架；使用 **GORM/GEN** 生成代码并操作 PostgreSQL 数据库
- 数据库：使用 **PostgreSQL** 和 **Redis** 集群实现读写分离，使用 **NebulaGraph** 存储用户关系
- 服务注册与发现：使用 **Consul** 作为服务发现与注册中心和配置中心
- 缓存策略：通过 **Write-Through** + **Write-Behind** 等多种策略提升数据访问速度。使用 **SingleFlight** 减轻数据库压力并防止缓存击穿、使用布隆过滤器减少缓存穿透，并通过随机延时策略避免缓存雪崩
- 中间件：采用**令牌桶**作为限流中间件，**JWT** 作为用户认证中间件，使用 **Kafka** 作为消息队列，实现异步写入数据库、配合 **Debezium** 删除缓存等操作
- 云原生：通过 **OpenTelemetry** + **Jaeger** 实现分布式链路追踪，**Prometheus** + **Grafana** 实现项目监控，使用 **Docker Compose** 一键部署项目，并通过 **GitHub Actions** 自动构建和推送镜像
- 其他：使用 **Snowflake** 算法生成全局唯一ID，使用 **ffmpeg** 截取视频第5帧作为封面，使用 **OSS** 存储视频和视频封面
## 代码生成示例:
```shell
1. Gorm/Gen代码生成
go run cmd/gen/generator.go
2. Kitex代码生成
cd src/service/feed
kitex -module douyin -service user -gen-path ../../kitex_gen/ ../../idl/user.thrift
```
## 未来规划
脱离青训营抖音客户端限制，对接口参数优化，添加新接口
* 添加双token
* 发布列表，喜欢列表，评论列表，关注列表，粉丝列表，好友列表增加分页参数
* 添加热门视频列表，关注视频列表，好友视频列表
* 评论增加点赞和回复功能，评论分为最新和最热列表，使用redis zset实现
* 聊天使用websocket