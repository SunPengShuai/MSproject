# 秒杀商城

## 目录结构
/MSproject
├── /images                        # 图片资源
├── /cmd
│   ├── /user-service              # 用户服务
│   ├── /order-service             # 订单服务
│   ├── /product-service           # 商品服务
│   ├── /stock-service             # 库存服务
│   ├── /page-service              # 页面基础服务
│   ├── /gateway-service          # API Gateway (基于 Kong 或自定义)
├── /internal
│   ├── /user                      # 用户相关业务逻辑
│   │   ├── /handler               # HTTP 请求处理器
│   │   ├── /service               # 业务逻辑服务
│   │   ├── /repository            # 数据库操作
│   │   ├── /model                 # 数据模型
│   │   └── /util                  # 工具函数
│   ├── /order                     # 订单相关业务逻辑
│   │   ├── /handler               # HTTP 请求处理器
│   │   ├── /service               # 业务逻辑服务
│   │   ├── /repository            # 数据库操作
│   │   ├── /model                 # 数据模型
│   │   └── /util                  # 工具函数
│   ├── /product                   # 商品相关业务逻辑
│   │   ├── /handler               # HTTP 请求处理器
│   │   ├── /service               # 业务逻辑服务
│   │   ├── /repository            # 数据库操作
│   │   ├── /model                 # 数据模型
│   │   └── /util                  # 工具函数
│   ├── /stock                     # 库存相关业务逻辑
│   │   ├── /handler               # HTTP 请求处理器
│   │   ├── /service               # 业务逻辑服务
│   │   ├── /repository            # 数据库操作
│   │   ├── /model                 # 数据模型
│   │   └── /util                  # 工具函数
│   ├── /page                      # 页面基础服务相关业务逻辑
│   │   ├── /handler               # HTTP 请求处理器
│   │   ├── /service               # 业务逻辑服务
│   │   ├── /repository            # 数据库操作
│   │   ├── /model                 # 数据模型
│   │   └── /util                  # 工具函数
│   ├── /common                    # 公共模块，如日志、配置、工具、错误处理等
│   │   ├── /config                # 配置文件加载
│   │   ├── /logger                # 日志处理
│   │   ├── /errors                # 错误处理
│   │   ├── /metrics               # 监控与度量
│   │   ├── /middleware            # 中间件（如认证、权限、日志等）
│   │   └── /util                  # 工具函数
├── /api
│   ├── /user                      # 用户服务API定义（Protobuf 或 HTTP）
│   ├── /order                     # 订单服务API定义（Protobuf 或 HTTP）
│   ├── /product                   # 商品服务API定义（Protobuf 或 HTTP）
│   ├── /stock                     # 库存服务API定义（Protobuf 或 HTTP）
│   ├── /page                      # 页面基础服务API定义（Protobuf 或 HTTP）
├── /scripts
│   ├── /migrations                # 数据库迁移脚本
│   ├── /deploy                    # 部署相关脚本
│   └── /build                     # 编译与构建脚本
├── /docker
│   ├── /user-service.Dockerfile   # 用户服务的 Dockerfile
│   ├── /order-service.Dockerfile  # 订单服务的 Dockerfile
│   ├── /product-service.Dockerfile# 商品服务的 Dockerfile
│   ├── /stock-service.Dockerfile  # 库存服务的 Dockerfile
│   ├── /gateway-service.Dockerfile# API Gateway 服务的 Dockerfile
│   └── /docker-compose.yml        # Docker Compose 文件，用于编排各个微服务
├── /deploy
│   ├── /k8s                       # Kubernetes 配置文件（Deployment、Service 等）
│   └── /helm                      # Helm charts 用于简化 Kubernetes 部署
├── /docs                           # 项目文档（功能、API、架构等）
├── /test                           # 测试相关代码
│   ├── /user-service              # 用户服务的测试
│   ├── /order-service             # 订单服务的测试
│   ├── /product-service           # 商品服务的测试
│   └── /common                    # 公共功能的测试
└── README.md

## 技术栈一览
 - etcd 服务注册和发现
 - kong 负载均衡和流量管理
 - grpc 通讯协议和rpc调用
 - redis 缓存信息
 - mysql / mongodb 数据库存储
 - rabbitmq 消息队列-削峰、解耦
 - docker 容器化技术
 - 下一版本技术
   - GitLab-CICD
   - K8s 弹性扩展滚动更新高可用
   - elasticsearch + kibana 搜索引擎+可视化
   - 监控和日志系统
## 技术架构图
![img.png](images/img.png)

## 功能实现
前端技术
  - vue （商城app实现）
  - 三剑客 （简单的秒杀前端页面实现）
后端主要实现以下几个微服务模块：
  - 用户信息模块
  - 订单模块
  - 商品信息模块
  - 库存服务模块
  - 页面基础服务模块

## 系统优化策略
### 页面静态化
秒杀商品页面的信息尽量写死，防止多余请求。\
HTML页面上生成倒计时的时钟，直到 秒杀开始 的时候，页面自动刷新，于此同时后台程序在CDN节点上面 更新了JS文件的内容。\
此时 加载到最新的js文件，于是js文件生成 秒杀按钮，用户点击按钮之后，js就会向后台程序发出秒杀请求。
### CDN 前端资源优化

### 预扣库存
先扣除了库存，保证不超卖，然后异步生成用户订单 \
用户拿到了订单，不支付怎么办？我们都知道现在订单都有有效期，比如说用户五分钟内不支付，订单就失效了，订单一旦失效，就会加入新的库存，这也是现在很多网上零售企业保证商品不少卖采用的方案。
### 消息队列解耦
订单的生成是异步的,一般都会放到MQ这样的即时消费队列中处理,订单量比较少的情况下，生成订单非常快，用户几乎不用排队。

### redis缓存——本地扣库存
把一定的库存量分配到本地机器，直接在内存中减库存，然后按照之前的逻辑异步创建订单
本地缓存->redis分布式缓存->数据库

### 缓存击穿问题：分布式锁
![img.png](images/huancunjichuan.png)
未命中时先拿锁再访问数据库
### 缓存穿透问题：
布隆过滤器/缓存空对象/提前过滤非法查询商品id 
### 缓存雪崩问题：
设置随机TTL和使用redis集群方案