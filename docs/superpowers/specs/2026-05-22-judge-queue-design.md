# 判题消息队列设计规格

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。

**目标：** 引入 RabbitMQ 消息队列，将代码判题流程从同步改为异步，支持高并发提交场景下的系统吞吐量和稳定性。

**架构：** 生产者（CodeSubmit/ContestSubmit）将判题任务推入 RabbitMQ，进程内 goroutine 池消费并执行判题，结果写入 DB + Redis，前端通过轮询或 WebSocket 获取结果。

**技术栈：** Go + Gin + GORM + Redis + RabbitMQ (amqp091-go) + gorilla/websocket

---

## 1. 现状与问题

当前判题流程为同步模式：

```
用户提交 → CodeSubmit → JudgeCode(同步阻塞) → HTTP 响应返回结果
```

**瓶颈：**
- 每次提交阻塞 HTTP 连接直到所有测试用例执行完毕（`go run` 最长可达 MaxRuntime 毫秒）
- 高并发时大量 `go run` 进程同时堆积，耗尽 CPU/内存
- 比赛期间集中提交会导致请求超时或服务器崩溃

## 2. 异步架构

```
用户提交 → Producer → RabbitMQ → Consumer(goroutine 池) → JudgeCode → DB + Redis
              ↓                                              ↓
         立即返回 submit_id                           WebSocket 推送 / 轮询可查
         + status=0(排队中)
```

## 3. RabbitMQ 设计

### 3.1 连接配置

通过环境变量配置：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `RABBITMQ_HOST` | `127.0.0.1` | 主机 |
| `RABBITMQ_PORT` | `5672` | 端口 |
| `RABBITMQ_USER` | `guest` | 用户名 |
| `RABBITMQ_PASSWORD` | `guest` | 密码 |

### 3.2 Exchange & Queue

- **Exchange:** `judge` (type: direct, durable)
- **Queue 1:** `judge.submit` — 普通代码提交
- **Queue 2:** `judge.contest` — 比赛内提交
- **Routing key:** `submit` / `contest`
- 消息持久化（delivery mode = 2）
- 消费者手动 ACK（autoAck = false）
- 失败消息 NACK + requeue，最多重试 3 次（通过消息 header 中的 `x-retry-count` 控制）

### 3.3 消息体

```go
type JudgeTask struct {
    SubmitIdentity  string `json:"submit_identity"`   // 提交记录 ID（已创建，status=0）
    CodePath        string `json:"code_path"`          // 代码文件路径（Producer 已保存）
    ProblemIdentity string `json:"problem_identity"`
    UserIdentity    string `json:"user_identity"`
    IsContest       bool   `json:"is_contest"`
    ContestIdentity string `json:"contest_identity,omitempty"`
}
```

## 4. 数据模型变更

### 4.1 submit_basic 表

status 字段现有值不变，新增 0 表示排队中：

| status | 含义 |
|--------|------|
| 0 | 排队中（新增） |
| 1 | 答案正确 (AC) |
| 2 | 答案错误 (WA) |
| 3 | 运行超时 (TLE) |
| 4 | 运行超内存 (MLE) |
| 5 | 编译错误 (CE) |
| 6 | 无效代码 |

### 4.2 contest_submit 表

同上，status=0 表示排队中。

### 4.3 Redis 结果缓存

- Key: `judge:result:{submit_identity}`
- Value: JSON
  ```json
  {
    "status": 1,
    "score": 100,
    "msg": "答案正确",
    "passed": 5,
    "total": 5
  }
  ```
- TTL: 5 分钟
- 用途：轮询接口快速查询，避免每次查 DB

## 5. Consumer 设计 (`service/judge_consumer.go`)

### 5.1 消费者池

- 启动 N 个 goroutine 消费 `judge.submit` 和 `judge.contest` 队列
- N 由环境变量 `JUDGE_WORKERS` 控制，默认 5
- 应用启动时在 `main.go` 中调用 `service.StartJudgeConsumers()`

### 5.2 消费流程

```
1. 从队列取消息
2. 反序列化 JudgeTask
3. 读取代码文件 (CodePath)
4. 调用 JudgeCode(code, problemIdentity)
5. 写入 DB：
   - 更新 submit_basic / contest_submit 的 status 和 score
   - 更新 user_basic 和 problem_basic 计数
6. 写入 Redis：judge:result:{submit_identity}
7. 通知 WebSocket Hub 推送结果
8. ACK 消息
```

### 5.3 重试机制

- 消息 header 中携带 `x-retry-count`
- 首次消费时 count=0，NACK requeue 时 +1
- 超过 3 次直接 ACK 丢弃（可选：写入死信队列 `judge.dlq`）

## 6. WebSocket 设计 (`service/ws_hub.go`)

### 6.1 Hub 结构

```go
type Hub struct {
    mu      sync.RWMutex
    clients map[string]map[*WsConn]bool // user_identity → connections
}

type WsConn struct {
    conn   *websocket.Conn
    userIdentity string
    send   chan []byte
}
```

### 6.2 连接管理

- 路由: `GET /ws?token={jwt_token}`
- 通过 JWT token 鉴权，提取 user_identity
- 注册到 Hub.clients[userIdentity]
- 心跳保活: 30s ping/pong
- 断线时自动从 Hub 移除

### 6.3 推送消息格式

```json
{
  "type": "judge_result",
  "submit_identity": "uuid",
  "status": 1,
  "score": 100,
  "msg": "答案正确"
}
```

### 6.4 前端集成

- 提交代码时建立 WebSocket 连接（如未建立）
- 收到 `judge_result` 消息后更新对应的提交记录状态
- 轮询作为降级方案：WebSocket 未连接时，前端每 2 秒轮询 `GET /user/submit-result?identity=xxx`

## 7. API 变更

### 7.1 改动接口

**`POST /user/code-submit`**（改动）

- 改动前：同步判题，返回 status 和 msg
- 改动后：创建 submit 记录（status=0），推入 MQ，返回：
  ```json
  {"code": 200, "data": {"submit_identity": "xxx", "status": 0, "msg": "已提交，排队中"}}
  ```

**`POST /user/contest-submit`**（改动）

- 同上，创建 contest_submit 记录（status=0），推入 MQ

### 7.2 新增接口

**`GET /user/submit-result?identity=xxx`**（新增，需用户 JWT）

- 查询判题结果，先查 Redis，未命中查 DB
- 返回: `{"code": 200, "data": {"status": 1, "score": 100, "msg": "答案正确", "passed": 5, "total": 5}}`

**`GET /ws?token=xxx`**（新增，WebSocket 升级）

- WebSocket 连接端点

## 8. Docker 部署变更

`docker-compose.yml` 新增 RabbitMQ 服务：

```yaml
rabbitmq:
  image: rabbitmq:3-management
  ports:
    - "5672:5672"
    - "15672:15672"
  environment:
    RABBITMQ_DEFAULT_USER: guest
    RABBITMQ_DEFAULT_PASS: guest
```

管理界面: http://localhost:15672

## 9. 文件清单

### 新增文件

| 文件 | 职责 |
|------|------|
| `Models/rabbitmq.go` | RabbitMQ 连接初始化、Producer 封装 |
| `service/judge_consumer.go` | 判题消费者 goroutine 池 |
| `service/ws_hub.go` | WebSocket Hub + 连接管理 |
| `service/ws_handler.go` | WebSocket HTTP handler |

### 修改文件

| 文件 | 改动 |
|------|------|
| `service/code.go` | CodeSubmit 改为异步：创建记录 + 推 MQ + 立即返回 |
| `service/contest.go` | ContestSubmit 改为异步：同上 |
| `router/app.go` | 新增 `/ws` 和 `/user/submit-result` 路由 |
| `Models/init.go` | 新增 RabbitMQ 连接初始化调用 |
| `main.go` | 启动 JudgeConsumers |
| `docker-compose.yml` | 新增 rabbitmq 服务 |
| `frontend/src/OnlineJudge.vue` | 提交后显示排队状态 + WebSocket 连接 + 结果更新 |
| `.env` | 新增 RabbitMQ 环境变量 |

## 10. 兼容性

- `JudgeCode` 函数签名和逻辑不变，Consumer 复用现有判题逻辑
- 普通提交和比赛提交共用同一套异步机制
- submit_basic 和 contest_submit 的 status 字段向后兼容（0=排队中，1-6 不变）
- 前端同时支持轮询和 WebSocket，WebSocket 不可用时自动降级为轮询
