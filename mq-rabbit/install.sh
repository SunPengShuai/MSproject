#!/bin/bash

# 定义 RabbitMQ 用户名、密码和 Erlang Cookie
RABBITMQ_USER="guest"
RABBITMQ_PASS="guest"
ERLANG_COOKIE="secret_cookie"

# 创建 docker-compose.yml 文件
cat << EOF > docker-compose.yml
version: '3.8'

services:
  rabbitmq1:
    image: rabbitmq:3-management
    hostname: rabbitmq1
    environment:
      RABBITMQ_DEFAULT_USER: $RABBITMQ_USER
      RABBITMQ_DEFAULT_PASS: $RABBITMQ_PASS
    ports:
      - "15671:15672"  # Management UI
      - "5671:5672"    # RabbitMQ
    volumes:
      - ./rabbitmq1/data:/var/lib/rabbitmq
      - ./rabbitmq1/etc/rabbitmq:/etc/rabbitmq
    networks:
      - rabbitmq-net

  rabbitmq2:
    image: rabbitmq:3-management
    hostname: rabbitmq2
    environment:
      RABBITMQ_DEFAULT_USER: $RABBITMQ_USER
      RABBITMQ_DEFAULT_PASS: $RABBITMQ_PASS
    ports:
      - "15672:15672"  # Management UI
      - "5672:5672"    # RabbitMQ
    volumes:
      - ./rabbitmq2/data:/var/lib/rabbitmq
      - ./rabbitmq2/etc/rabbitmq:/etc/rabbitmq
    networks:
      - rabbitmq-net

  rabbitmq3:
    image: rabbitmq:3-management
    hostname: rabbitmq3
    environment:
      RABBITMQ_DEFAULT_USER: $RABBITMQ_USER
      RABBITMQ_DEFAULT_PASS: $RABBITMQ_PASS
    ports:
      - "15673:15672"  # Management UI
      - "5673:5672"    # RabbitMQ
    volumes:
      - ./rabbitmq3/data:/var/lib/rabbitmq
      - ./rabbitmq3/etc/rabbitmq:/etc/rabbitmq
    networks:
      - rabbitmq-net

networks:
  rabbitmq-net:
EOF

echo "docker-compose.yml 文件已创建。"

# 创建节点目录结构
for i in {1..3}; do
    mkdir -p rabbitmq$i/data rabbitmq$i/etc/rabbitmq
    echo "NODENAME=rabbit@rabbitmq$i" > ./rabbitmq$i/etc/rabbitmq/rabbitmq-env.conf
    echo "$ERLANG_COOKIE" > ./rabbitmq$i/etc/rabbitmq/.erlang.cookie
    chmod 400 ./rabbitmq$i/etc/rabbitmq/.erlang.cookie  # 设置权限为 400
done

echo "各节点的配置文件已创建。"

# 启动 Docker Compose
docker-compose up -d

# 等待 RabbitMQ 启动
sleep 10

# 将节点加入集群
docker exec -it rabbitmq2 rabbitmqctl stop_app
docker exec -it rabbitmq2 rabbitmqctl join_cluster rabbitMq@rabbitmq1
docker exec -it rabbitmq2 rabbitmqctl start_app

docker exec -it rabbitmq3 rabbitmqctl stop_app
docker exec -it rabbitmq3 rabbitmqctl join_cluster rabbitMq@rabbitmq1
docker exec -it rabbitmq3 rabbitmqctl start_app

# 显示集群状态
docker exec -it rabbitmq1 rabbitmqctl cluster_status

echo "RabbitMQ 集群已成功启动。"
