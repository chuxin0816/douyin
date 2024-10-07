#!/bin/sh

# 先执行默认的入口点脚本以启动 Debezium 服务
/docker-entrypoint.sh start &

# 等待 Debezium 连接器服务启动，直到返回 HTTP 状态码 200
while [ "$(curl -o /dev/null -s -w "%{http_code}" http://localhost:8083/)" != "200" ]; do
  echo "等待 Debezium 连接器服务启动..."
  sleep 5
done

echo "Debezium 连接器服务已启动"

# 配置 Debezium 连接器
curl -i -X POST -H "Accept:application/json" -H "Content-Type:application/json" --data @/etc/debezium/debezium.json http://localhost:8083/connectors/

# 保持容器运行
wait