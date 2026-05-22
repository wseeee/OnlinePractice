# 比赛系统设计规格

日期: 2026-05-21
状态: 待实现

## 概述

为 OnlineJudge 系统添加 OI 赛制的比赛功能，支持定时比赛和公开/报名制两种参赛模式。

## 赛制规则

- OI 赛制：每题满分 100 分，按通过测试用例比例计分
- 计分公式：`score = passed * ProblemScore / total`（整除取整，ProblemScore 默认 100）
- 排名：总分降序，同分按各题最早 AC 时间之和升序（无 AC 的题目不计入时间）
- 每用户每题取所有提交中的最高分

## 数据模型

### contest_basic — 比赛主表

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| Identity | varchar(36) | UUID |
| Title | varchar(255) | 比赛标题 |
| Description | text | 比赛描述 |
| StartAt | datetime | 开始时间 |
| EndAt | datetime | 结束时间 |
| ContestType | tinyint(1) | 1=公开, 2=报名制 |
| MaxParticipants | int | 报名制上限，0=不限 |
| CreatedAt | datetime | |
| UpdatedAt | datetime | |
| DeletedAt | datetime | 软删除 |

### contest_problem — 比赛-题目关联

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| ContestIdentity | varchar(36) | 关联比赛 |
| ProblemIdentity | varchar(36) | 关联题目 |
| Score | int | 满分值，默认 100 |
| Sort | int | 题目排序 |

### contest_user — 报名记录

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| ContestIdentity | varchar(36) | 关联比赛 |
| UserIdentity | varchar(36) | 关联用户 |
| CreatedAt | datetime | 报名时间 |

### contest_submit — 比赛提交记录

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| ContestIdentity | varchar(36) | 关联比赛 |
| ProblemIdentity | varchar(36) | 关联题目 |
| UserIdentity | varchar(36) | 关联用户 |
| Path | varchar(255) | 代码文件路径 |
| Score | int | 本次得分 |
| Status | tinyint(1) | 判题状态（复用现有 1-6） |
| CreatedAt | datetime | 提交时间 |

## API 设计

### 管理员接口 (/admin)

**POST /admin/contest-create**
- 参数: title, description, start_at, end_at, contest_type, max_participants, problem_identities(JSON 数组)
- 创建比赛并关联题目

**PUT /admin/contest-update**
- 参数: identity(query), 同上(form)
- 修改比赛信息，事务内更新关联题目

**GET /admin/contest-list**
- 参数: page, size, keyword, status(0=全部, 1=未开始, 2=进行中, 3=已结束)
- 比赛管理列表

**DELETE /admin/contest-delete**
- 参数: identity(query)
- 软删除比赛

### 用户接口 (/user)

**POST /user/contest-join**
- 参数: contest_identity(query)
- 报名比赛，校验：比赛存在、报名制、未满员、未重复报名

**POST /user/contest-submit**
- 参数: contest_identity(query), problem_identity(query), code(body)
- 比赛内提交代码
- 校验：比赛存在、当前时间在比赛期间、报名制需已报名
- 复用 JudgeCode 判题，计算 OI 分数，写入 contest_submit

**GET /user/contest-submits**
- 参数: contest_identity(query), problem_identity(query 可选)
- 查看自己在某比赛的提交记录

### 公共接口

**GET /contest-list**
- 参数: page, size
- 比赛列表，按开始时间倒序

**GET /contest-detail**
- 参数: identity(query)
- 比赛详情，含题目列表（标题、满分）、时间、参赛状态

**GET /contest-rank**
- 参数: contest_identity(query), page, size
- 比赛排行榜，Redis 缓存 30 秒

## 业务逻辑

### 判题复用

从 `service/code.go` 提取公共函数：

```go
func JudgeCode(code []byte, problemIdentity string) (passed, total int, status int, err error)
```

- `CodeSubmit` 调用后写入 submit_basic
- `ContestSubmit` 调用后计算分数写入 contest_submit

### 排名计算

1. 查询 contest_submit 中该比赛所有记录
2. 按 (user_identity, problem_identity) 分组，每组取 max(score)，同时记录该题最早 AC(状态=1) 的时间
3. 按用户聚合：总分 = 各题最高分之和，总时间 = 各题最早 AC 时间之和
4. 按总分降序，同分按总时间升序
5. 结果缓存到 Redis（key: `contest_rank:{identity}`，TTL 30s）

### 比赛状态

- 未开始: now < StartAt
- 进行中: StartAt <= now <= EndAt
- 已结束: now > EndAt

## 前端

### 新增视图

- `contest` — 比赛列表，展示标题、时间、状态标签、参赛人数
- `contest-detail` — 比赛详情页：题目列表（带得分）、代码编辑器、排行榜

### 管理员面板

新增比赛管理 tab：创建/编辑表单、题目选择、时间设置、参赛模式选择。

## 文件结构

```
Models/contest_basic.go         比赛模型 + 数据访问
Models/contest_problem.go       比赛-题目关联模型
Models/contest_user.go          报名记录模型
Models/contest_submit.go        比赛提交模型
service/contest.go              比赛相关 HTTP handler
router/app.go                   新增路由注册
frontend/src/views/OnlineJudge.vue  新增 contest/contest-detail 视图
```

## 不做的事情

- 不支持 ACM 赛制（仅 OI）
- 不支持 Special Judge
- 不支持比赛内实时 WebSocket 推送（用轮询 + Redis 缓存替代）
- 不支持虚拟比赛（随时参加模式）
