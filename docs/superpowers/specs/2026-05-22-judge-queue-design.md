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
         立即返回 submit_identity                       WebSocket 推送 / 轮询可查
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
- **Queue 1:** `judge.submit` — 普通代码提交 (durable)
- **Queue 2:** `judge.contest` — 比赛内提交 (durable)
- **Dead Letter Exchange:** `judge.dlx` (type: direct, durable)
- **Dead Letter Queue:** `judge.dlq` (durable)
- **Routing key:** `submit` / `contest`
- 消息持久化（delivery mode = 2）
- 消费者手动 ACK（autoAck = false）

### 3.3 重试机制（死信队列 + TTL 延迟重试）

不使用 NACK+requeue（会导致消息重回队头，无延迟，易死循环）。采用死信队列延迟重试：

```
正常队列 judge.submit
  → 消费失败 → ACK + 发布到 judge.dlx（消息 header 携带 x-retry-count）
    → judge.dlq（TTL 30 秒后过期，重新路由回 judge.submit）
      → 重试消费，x-retry-count + 1
        → 超过 3 次 → 从 judge.dlq 丢弃（ACK，不 requeue）
```

消息 header 字段：
- `x-retry-count`: int，初始 0，每次进死信队列 +1
- `x-original-queue`: string，原始队列名（submit 或 contest）

### 3.4 消息体

```go
type JudgeTask struct {
    SubmitIdentity  string `json:"submit_identity"`   // 提交记录 ID（已创建，status=0）
    CodePath        string `json:"code_path"`          // 代码文件路径（Producer 已保存）
    ProblemIdentity string `json:"problem_identity"`
    UserIdentity    string `json:"user_identity"`
    IsContest       bool   `json:"is_contest"`
    ContestIdentity string `json:"contest_identity,omitempty"`
    ProblemScore    int    `json:"problem_score,omitempty"` // 比赛题分值（仅比赛提交）
}
```

**注意：** `ProblemScore` 字段用于比赛提交的 OI 计分。Producer 在创建消息时从 `contest_problem` 表读取分值并填入，Consumer 判题后直接使用，无需再查库。

## 4. 数据模型变更

### 4.1 contest_submit 表 — 新增 identity 字段

当前 `ContestSubmit` 没有 `Identity` 字段，无法作为 Redis 缓存键和 WebSocket 推送的主键。

**新增字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| identity | VARCHAR(36) | 提交唯一标识（UUID），用于 MQ 消息关联、Redis 缓存键、WebSocket 推送 |

**GORM 模型变更：**
```go
type ContestSubmit struct {
    gorm.Model
    Identity        string        `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
    ContestIdentity string        `gorm:"column:contest_identity;type:varchar(36);" json:"contest_identity"`
    // ... 其余字段不变
}
```

**数据库迁移：**
```sql
ALTER TABLE `contest_submit` ADD COLUMN `identity` VARCHAR(36) NOT NULL AFTER `id`;
ALTER TABLE `contest_submit` ADD UNIQUE INDEX `idx_contest_submit_identity` (`identity`);
```

### 4.2 submit_basic 表 — status=0 语义

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

### 4.3 GetSubmitList 查询修复

当前 `Models/submit_basic.go:37` 使用 `if status != 0` 作为"不筛选"的条件。status=0 变为有效值后必须修复：

```go
// 修复前：status=0 意味着"不筛选"
if status != 0 {
    tx = tx.Where("status = ?", status)
}

// 修复后：用 -1 表示"不筛选"，0 是有效值"排队中"
if status >= 0 {
    tx = tx.Where("status = ?", status)
}
```

**调用方变更：** 查询全部状态时传 `status = -1` 而非 `0`。

### 4.4 Redis 结果缓存

- Key: `judge:result:{submit_identity}`（submit_basic 和 contest_submit 统一使用 identity 字段）
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

## 5. JudgeCode 重构 — 代码保存职责分离

### 5.1 问题

当前 `JudgeCode` 函数内部调用 `Helper.SaveCode(code)` 保存代码文件。异步化后 Producer 需要先保存代码（拿到 CodePath 才能序列化到消息中），如果 JudgeCode 内部再保存一次就会产生重复文件。

### 5.2 方案

将 `JudgeCode` 拆分为两个函数：

```go
// SaveAndValidateCode — Producer 调用：保存代码 + 校验合法性
// 返回 codePath 和 error。失败时无需入队。
func SaveAndValidateCode(code []byte, problemIdentity string) (codePath string, err error) {
    path, err := Helper.SaveCode(code)
    if err != nil {
        return "", err
    }
    v, err := Helper.CheckGoCodeValid(path)
    if err != nil {
        return "", err
    }
    if !v {
        return path, nil // 返回 path 但 Consumer 会识别 status=6
    }
    return path, nil
}

// JudgeCode — Consumer 调用：纯判题（不保存代码）
// 签名变更：接受 codePath（string）而非 code（[]byte）
func JudgeCode(codePath string, problemIdentity string) (passed, total int, status int, err error) {
    // 不再调用 Helper.SaveCode
    // 直接使用 codePath 读取代码、校验、执行测试用例
    // ... 其余逻辑不变
}
```

**关键变更：**
- `JudgeCode` 不再调用 `Helper.SaveCode`，不再返回 `path`
- Producer 调用 `SaveAndValidateCode` 保存代码，失败直接返回错误
- Consumer 从消息中取 `CodePath`，传入 `JudgeCode` 执行判题

## 6. Consumer 设计 (`service/judge_consumer.go`)

### 6.1 消费者池

- 启动 N 个 goroutine 消费 `judge.submit` 和 `judge.contest` 队列
- N 由环境变量 `JUDGE_WORKERS` 控制，默认 5
- 应用启动时在 `main.go` 中调用 `service.StartJudgeConsumers()`

### 6.2 普通提交消费流程

```
1. 从 judge.submit 取消息
2. 反序列化 JudgeTask
3. 调用 JudgeCode(codePath, problemIdentity)
4. 写入 DB：
   a. 更新 submit_basic 的 status
   b. 更新 user_basic 的 submit_num（+1）和 pass_num（AC 时 +1）
   c. 更新 problem_basic 的 submit_num（+1）和 pass_num（AC 时 +1）
5. 写入 Redis：judge:result:{submit_identity}
6. 通知 WebSocket Hub 推送结果
7. ACK 消息
```

### 6.3 比赛提交消费流程

```
1. 从 judge.contest 取消息
2. 反序列化 JudgeTask（含 ProblemScore）
3. 调用 JudgeCode(codePath, problemIdentity)
4. OI 计分：score = passed * task.ProblemScore / total
5. 写入 DB：
   a. 更新 contest_submit 的 status 和 score
   b. 更新 user_basic 的 submit_num（+1）和 pass_num（AC 时 +1）
   c. 更新 problem_basic 的 submit_num（+1）和 pass_num（AC 时 +1）
6. 失效 Redis 排行榜缓存：DEL contest_rank:{contest_identity}
7. 写入 Redis：judge:result:{submit_identity}
8. 通知 WebSocket Hub 推送结果
9. ACK 消息
```

**比赛计分注意：**
- `ProblemScore` 由 Producer 在入队时从 `contest_problem` 表读取并填入消息体
- Consumer 判题后直接用 `passed * ProblemScore / total` 计算，无需再查 contest_problem 表
- 每次比赛提交完成后，删除 `contest_rank:{contest_identity}` 缓存，下次查询排行榜时自动重建

### 6.4 重试流程

```
消费失败 → ACK（从正常队列移除）
         → 发布到 judge.dlx，header 中 x-retry-count + 1
           → judge.dlq（TTL 30s）
             → 过期后路由回 judge.submit
               → 重新消费
                 → x-retry-count >= 3 → 从 judge.dlq ACK 丢弃
```

## 7. WebSocket 设计 (`service/ws_hub.go`)

### 7.1 Hub 结构

```go
type Hub struct {
    mu      sync.RWMutex
    clients map[string]map[*WsConn]bool // user_identity → connections
}

type WsConn struct {
    conn         *websocket.Conn
    userIdentity string
    send         chan []byte
}
```

### 7.2 连接管理

- 路由: `GET /ws?token={jwt_token}`
- 通过 JWT token 鉴权，提取 user_identity
- 注册到 Hub.clients[userIdentity]
- 心跳保活: 30s ping/pong
- 断线时自动从 Hub 移除

### 7.3 推送消息格式

```json
{
  "type": "judge_result",
  "submit_identity": "uuid",
  "status": 1,
  "score": 100,
  "msg": "答案正确"
}
```

### 7.4 前端集成

- 提交代码时建立 WebSocket 连接（如未建立）
- 收到 `judge_result` 消息后更新对应的提交记录状态
- 轮询作为降级方案：WebSocket 未连接时，前端每 2 秒轮询 `GET /user/submit-result?identity=xxx`

## 8. API 变更

### 8.1 改动接口

**`POST /user/code-submit`**（改动）

- 改动前：同步判题，返回 status 和 msg
- 改动后：
  1. 调用 `SaveAndValidateCode` 保存代码，失败直接返回错误
  2. 创建 submit 记录（status=0，path=codePath）
  3. 构造 JudgeTask 推入 RabbitMQ
  4. 立即返回：
     ```json
     {"code": 200, "data": {"submit_identity": "xxx", "status": 0, "msg": "已提交，排队中"}}
     ```

**`POST /user/contest-submit`**（改动）

- 改动前：同步判题 + OI 计分，返回 score 和 status
- 改动后：
  1. 校验比赛状态、报名状态、题目归属（同现有逻辑）
  2. 调用 `SaveAndValidateCode` 保存代码
  3. 创建 contest_submit 记录（status=0，identity=UUID，path=codePath）
  4. 从 `contest_problem` 表读取 `Score` 字段（分值）
  5. 构造 JudgeTask（含 ProblemScore）推入 RabbitMQ
  6. 立即返回：
     ```json
     {"code": 200, "data": {"submit_identity": "xxx", "status": 0, "msg": "已提交，排队中"}}
     ```

### 8.2 新增接口

**`GET /user/submit-result?identity=xxx`**（新增，需用户 JWT）

- 先查 Redis `judge:result:{identity}`，命中直接返回
- 未命中查 DB（submit_basic 或 contest_submit，根据 identity 前缀或同时查两张表）
- 返回:
  - 普通提交: `{"code": 200, "data": {"status": 1, "msg": "答案正确"}}`
  - 比赛提交: `{"code": 200, "data": {"status": 1, "score": 85, "msg": "答案正确"}}`

**`GET /ws?token=xxx`**（新增，WebSocket 升级）

- WebSocket 连接端点

## 9. Docker 部署变更

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

## 10. 文件清单

### 新增文件

| 文件 | 职责 |
|------|------|
| `Models/rabbitmq.go` | RabbitMQ 连接初始化、Exchange/Queue 声明、Producer 封装 |
| `service/judge_consumer.go` | 判题消费者 goroutine 池 + 重试逻辑 |
| `service/ws_hub.go` | WebSocket Hub + 连接管理 |
| `service/ws_handler.go` | WebSocket HTTP handler（鉴权 + 升级） |

### 修改文件

| 文件 | 改动 |
|------|------|
| `service/code.go` | 拆分 JudgeCode；CodeSubmit 改为异步：SaveAndValidateCode + 创建记录 + 推 MQ |
| `service/contest.go` | ContestSubmit 改为异步：同上 + 读取 ProblemScore + 创建带 identity 的记录 |
| `Models/contest_submit.go` | 新增 Identity 字段 |
| `Models/submit_basic.go` | GetSubmitList 查询修复（status=0 语义） |
| `router/app.go` | 新增 `/ws` 和 `/user/submit-result` 路由 |
| `Models/init.go` | 新增 RabbitMQ 连接初始化调用 |
| `main.go` | 启动 JudgeConsumers |
| `docker-compose.yml` | 新增 rabbitmq 服务 |
| `migrations/002_judge_queue.sql` | contest_submit 添加 identity 字段 |
| `frontend/src/OnlineJudge.vue` | 提交后显示排队状态 + WebSocket 连接 + 结果更新 |
| `.env` | 新增 RabbitMQ 环境变量 |

## 11. 兼容性

- 普通提交和比赛提交共用同一套异步机制
- submit_basic 和 contest_submit 的 status 字段向后兼容（0=排队中，1-6 不变）
- 前端同时支持轮询和 WebSocket，WebSocket 不可用时自动降级为轮询
- `GetSubmitList` 查询语义变更：调用方需传 -1 表示"全部状态"，0 表示筛选"排队中"
- 比赛排行榜 Redis 缓存在每次比赛提交后自动失效
