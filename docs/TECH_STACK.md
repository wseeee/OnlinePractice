# OnlineJudge 后端技术栈总结

## 项目定位

基于 Go 语言的在线判题系统（Online Judge）后端服务，支持用户注册登录、题库管理、Go 代码在线提交与自动判题、排行榜等核心功能。前后端分离架构，后端提供 RESTful API，约 2400 行业务代码。

---

## 一、基础框架

| 组件 | 选型 | 版本 | 说明 |
|------|------|------|------|
| 语言 | Go | 1.26 | 模块名 `OnlinePrictice` |
| HTTP 框架 | **Gin** | v1.12.0 | 路由、中间件、参数绑定、JSON 响应 |
| ORM | **GORM** | v1.31.1 | MySQL ORM，含自动迁移、Preload 预加载、事务 |
| 数据库驱动 | gorm.io/driver/mysql | v1.6.0 | GORM MySQL 驱动 |
| Redis 客户端 | **go-redis** | v8.11.5 | 验证码临时存储，5 分钟 TTL |

### 选型说明

- **Gin**：Go 生态中最成熟、性能最高的 HTTP 框架，社区活跃，中间件生态丰富，适合构建 RESTful API
- **GORM**：Go 最流行的 ORM，通过 Preload 机制处理关联查询，支持事务和软删除，适合中小规模数据层
- **go-redis**：Redis 官方推荐的 Go 客户端，支持连接池、Pipeline、Pub/Sub，满足缓存与临时数据存储需求

---

## 二、认证与鉴权

| 组件 | 选型 | 版本 | 说明 |
|------|------|------|------|
| JWT 库 | **golang-jwt/jwt** | v5.3.1 | Token 签发与解析，HS256 签名 |
| 密码处理 | crypto/md5 | 标准库 | MD5 哈希存储密码 |

### 鉴权架构

```
请求 → Gin Router
      ├─ 公共路由（无需认证）
      │   ├─ 用户注册/登录、发送验证码
      │   ├─ 题库列表/详情、排行榜、提交记录
      │   └─ Swagger 文档
      ├─ /user/*  → AuthUserCheck 中间件 → 验证 JWT + IsAdmin == 0
      └─ /admin/* → AuthAdminCheck 中间件 → 验证 JWT + IsAdmin == 1
```

### 鉴权流程

1. 用户登录/注册成功后，后端签发 JWT Token（载荷含 `identity`、`name`、`is_admin`）
2. 客户端将 Token 存入 `Authorization` 请求头
3. 中间件拦截请求，调用 `AnalyseToken` 解析并验证 Token 合法性
4. 验证通过后将 `UserJwt` 结构体注入 `gin.Context`，后续 handler 通过 `c.Get("user_claims")` 获取
5. 管理员路由额外校验 `IsAdmin == 1`，实现用户/管理员权限隔离

---

## 三、数据存储

### MySQL（主存储）

**数据库**：`onlinepractice` · **字符集**：utf8mb4 · **引擎**：InnoDB

| 表名 | 对应模型 | 说明 |
|------|----------|------|
| `user_basic` | Models.UserBasic | 用户表（identity, name, password, mail, phone, pass_num, submit_num, is_admin） |
| `problem_basic` | Models.ProblemBasic | 题目表（identity, title, content, max_mem, max_runtime, submit_num, pass_num） |
| `category_basic` | Models.CategoryBasic | 分类表（identity, name, parent_id），支持树形层级 |
| `problem_category` | Models.ProblemCategory | 题目-分类关联表（problem_id, category_id），多对多 |
| `test_case` | Models.TestCase | 测试用例表（identity, problem_identity, input, output），一对多 |
| `submit_basic` | Models.SubmitBasic | 提交记录表（identity, problem_identity, user_identity, path, status） |

**ORM 特性应用**：
- **软删除**：所有模型内嵌 `gorm.Model`（ID, CreatedAt, UpdatedAt, DeletedAt）
- **关联预加载**：`Preload("ProblemCategories").Preload("TestCases")` 解决 N+1 查询
- **事务**：题目更新和代码提交流程使用 `DB.Transaction` 保证原子性
- **原子计数**：`gorm.Expr("submit_num + ?", 1)` 实现并发安全的计数更新

### Redis（缓存）

- **用途**：邮箱验证码临时存储，Key 为邮箱地址，Value 为 6 位数字验证码
- **过期策略**：`SET key value EX 300`，5 分钟自动过期
- **部署**：独立虚拟机实例（当前硬编码连接地址）

---

## 四、业务组件

### 用户系统

```
注册流程：填写表单 → 邮箱验证码（存入 Redis）→ 验证码校验 → MD5 加密密码 → 写入 MySQL → 签发 JWT
登录流程：用户名+密码 → MD5 密码比对 → 签发 JWT
```

| 组件 | 文件 | 说明 |
|------|------|------|
| 邮箱发送 | Helper/SendEmail | SMTP 连接 QQ 邮箱（smtp.qq.com:587），发送 HTML 格式验证码 |
| 验证码生成 | Helper/GenerateCode | `crypto/rand` 生成 6 位随机数字 |
| UUID 生成 | Helper/GetUUID | `satori/go.uuid` 生成 UUID v4 作为资源唯一标识 |
| JWT 签发 | Helper/GenerateToken | HS256 签名，密钥 `Online_Practice`，载荷含用户身份信息 |
| JWT 解析 | Helper/AnalyseToken | 解析并验证 Token，返回 `UserJwt` 结构体 |

### 代码判题引擎（核心业务）

```
用户提交 Go 代码
  → SaveCode：写入 code/<uuid>/main.go
  → CheckGoCodeValid：校验 import 白名单（bytes, fmt, math, sort, strings）
  → 遍历测试用例，并发执行 go run <path>
  → stdin 传入测试输入，捕获 stdout/stderr
  → 判定维度：
     ├─ 答案正确 (AC)：stdout == expected output 且 passCount == len(testCases)
     ├─ 答案错误 (WA)：stdout != expected output
     ├─ 运行超时 (TLE)：time.After(MaxRuntime) 触发
     ├─ 内存超限 (MLE)：runtime.ReadMemStats 前后差值 > MaxMem
     ├─ 编译错误 (CE)：exit status 2
     └─ 无效代码 (EC)：import 不在白名单内
  → goroutine + channel + select 多路监听，任一判定触发即结束
  → DB.Transaction 写入 submit_basic + 原子更新 user_basic / problem_basic 计数
```

**并发模型**：
- 每个测试用例启动一个 goroutine 并行执行
- 使用 `sync.Mutex` 保护 `passCount` 共享变量
- 使用 5 个 channel（AC/WA/TLE/MLE/CE）+ 1 个 EC channel 进行结果收集
- `select` 多路复用，取最先到达的结果作为最终判定

### 分页查询模式

所有列表接口采用统一分页约定：
- 默认 `page=1`，`size=20`
- 数据库层 `Offset((page-1)*size).Limit(size)`
- 响应格式 `{"code":200, "data":{"list":[...], "count": N}}`

---

## 五、中间件

| 中间件 | 文件 | 功能 |
|--------|------|------|
| AuthUserCheck | middleware/auth_user.go | 解析 JWT，校验用户身份（IsAdmin == 0），注入 user_claims 到 Context |
| AuthAdminCheck | middleware/auth_admin.go | 解析 JWT，校验管理员身份（IsAdmin == 1），拦截非管理员请求 |
| gin.Default() | 内置 | Logger + Recovery 中间件，自动处理 panic 恢复和请求日志 |

---

## 六、API 文档

| 组件 | 选型 | 说明 |
|------|------|------|
| Swagger 生成 | swaggo/swag | 通过 Go 注释生成 OpenAPI 2.0 规范文档 |
| Swagger UI | swaggo/gin-swagger | 在 `/swagger/index.html` 提供可视化接口文档 |

---

## 七、辅助依赖

| 包 | 用途 |
|----|------|
| crypto/md5 | 密码哈希 |
| crypto/rand | 验证码随机数生成 |
| net/smtp | SMTP 邮件发送 |
| encoding/json | 测试用例 JSON 解析 |
| io/ioutil | 文件读写（代码保存与读取） |
| os/exec | 执行 `go run` 命令运行用户代码 |
| runtime | 读取内存统计（判题内存判定） |
| sync | goroutine 互斥锁 |

---

## 八、整体架构图

```
┌─────────────────────────────────────────────────────────┐
│                     Gin HTTP Server (:8080)              │
│                                                         │
│  ┌─────────────────┐  ┌────────────────────────────┐    │
│  │   Public Routes  │  │   Protected Routes          │    │
│  │                  │  │  ┌──────────┐ ┌──────────┐ │    │
│  │  /login          │  │  │ /user/*  │ │ /admin/* │ │    │
│  │  /register       │  │  │  AuthUser │ │AuthAdmin │ │    │
│  │  /sendcode       │  │  └────┬─────┘ └────┬─────┘ │    │
│  │  /problem-*      │  │       │            │       │    │
│  │  /rank-list      │  │       ▼            ▼       │    │
│  │  /submit-list    │  │  ┌─────────────────────┐   │    │
│  │  /user-basic     │  │  │   JWT Middleware     │   │    │
│  │  /swagger/*      │  │  └─────────┬───────────┘   │    │
│  └────────┬─────────┘  └────────────┼────────────────┘    │
│           │                         │                     │
│           └──────────┬──────────────┘                     │
│                      ▼                                    │
│          ┌───────────────────────┐                        │
│          │    Service Layer      │                        │
│          │  (user/problem/code/  │                        │
│          │   submit/category)    │                        │
│          └───────────┬───────────┘                        │
│                      │                                    │
│     ┌────────────────┼────────────────┐                   │
│     ▼                ▼                ▼                   │
│  ┌──────┐    ┌───────────┐    ┌──────────┐               │
│  │Helper│    │  Models   │    │  Define  │               │
│  │ JWT  │    │  GORM     │    │  Consts  │               │
│  │Email │    │  Models   │    │  Types   │               │
│  │UUID  │    └─────┬─────┘    └──────────┘               │
│  │Code  │          │                                      │
│  └──────┘          ▼                                      │
│             ┌──────────────────┐                          │
│             │  MySQL  │  Redis │                          │
│             └──────────────────┘                          │
└─────────────────────────────────────────────────────────┘
```

---

## 九、项目统计

| 维度 | 数据 |
|------|------|
| Go 业务代码行数 | ~2400 行（不含 docs.go 自动生成） |
| 模块数 | 7（router, middleware, service, Models, Helper, define, docs） |
| API 端点 | 15 个（10 公共 + 1 用户 + 4 管理员） |
| 数据表 | 6 张 |
| 测试文件 | 5 个（GORM 连接、JWT、Email、Redis、UUID） |
| 直接依赖 | 8 个（gin, gorm, go-redis, jwt, email, uuid, swaggo） |

---

> 最后更新：2026-05-13 · Go 1.26 · Gin v1.12.0 · GORM v1.31.1
