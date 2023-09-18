# 基于Hertz框架的WEB开发脚手架

包含了常用的WEB开发组件的初始化，并基于CLD架构分层

1. 使用viper读取config.yaml中的配置

2. 使用gorm和go-redis连接mysql和redis数据库

3. 使用zap(Hertz)和lumberjack记录和按日期保存日志

4. 使用Hertz启动服务，实现路由和优雅关机等操作
