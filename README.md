# 抖音
> 字节跳动青训营项目，简易版抖音，实现了抖音的基本功能
## 接口文档：https://apifox.com/apidoc/shared-119368ca-740a-467c-99f3-f7ca31df29c2
## 项目演示地址：http://chuxin0816.com:8888/ (已关闭)
### 项目部署
项目使用docker-compose部署，在配置好config/config.yaml后运行docker-compose up即可启动项目
******
## 技术选型：
* 使用Hertz作为http微服务框架，具有高性能，高可用，高扩展性的特点
* 使用GORM操作Mysql，具有简单易用，防SQL注入等优点
* 使用Redis作为缓存数据库，提高访问速度，使用延时双删保证数据一致性
* 使用布隆过滤器防止缓存穿透，使用随机延迟防止缓存雪崩
* 使用SingleFlight防止缓存击穿
* 使用hertz集成的zap日志库记录日志
* 使用JWT作为用户认证，使用中间件进行认证
* 使用bcrypt加密用户密码
* 使用snowflake生成各种id，使用uuid生成oss文件名
* 使用阿里云oss存储视频文件
* 使用ffmpeg进行视频转码
* 使用令牌桶作为限流中间件
* 使用viper读取配置文件并用air进行热更新
* 使用docker-compose部署项目
