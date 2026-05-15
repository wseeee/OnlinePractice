# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

<!-- superpowers-zh:begin (do not edit between these markers) -->
# Superpowers-ZH 中文增强版

本项目已安装 superpowers-zh 技能框架（20 个 skills）。

## 核心规则

1. **收到任务时，先检查是否有匹配的 skill** — 哪怕只有 1% 的可能性也要检查
2. **设计先于编码** — 收到功能需求时，先用 brainstorming skill 做需求分析
3. **测试先于实现** — 写代码前先写测试（TDD）
4. **验证先于完成** — 声称完成前必须运行验证命令

## 如何使用

当任务匹配某个 skill 时，使用 `Skill` 工具加载对应 skill 并严格遵循其流程。绝不要用 Read 工具读取 SKILL.md 文件。
<!-- superpowers-zh:end -->

## 项目概述

OnlineJudge 在线判题系统 — Go + Gin 后端，Vue 3 前端，MySQL + Redis 存储。

## 常用命令

```bash
# 后端 — 在项目根目录运行
go run main.go                           # 启动后端 (端口 8080)
go test ./test/...                       # 运行所有测试 (需要 MySQL + Redis)
swag init                                # 重新生成 Swagger 文档

# 前端 — 在 frontend/ 目录运行
npm run dev                              # Vite 开发服务器 (端口 3000)
npm run build                            # 生产构建
```

## 技术栈与依赖

- **Go 1.26**, module: `OnlinePrictice`
- **Gin** — HTTP 框架
- **GORM** (MySQL driver) — ORM, 数据库 `onlinepractice` @ 127.0.0.1:3306
- **go-redis/v8** — Redis 缓存 (验证码存储，TTL 5 分钟)
- **golang-jwt/v5** — JWT 认证, 签名密钥 `Online_Practice`
- **jordan-wright/email** — QQ 邮箱 SMTP 发送验证码
- **swaggo** — Swagger 文档自动生成，访问 `/swagger/index.html`
- **uuid (satori/go.uuid)** — 生成 UUID v4 作为资源 identity
- **Vue 3 + Vite + Axios** — 前端单页应用

## 架构分层

```
main.go                         入口，调用 router.Router().Run()
router/app.go                   注册所有路由 (public / user / admin 三组)
middleware/auth_user.go         用户 JWT 验证中间件 (IsAdmin == 0)
middleware/auth_admin.go        管理员 JWT 验证中间件 (IsAdmin == 1)
service/*.go                    HTTP handler，参数解析 → Model 查询 → JSON 响应
Models/*.go                     GORM 模型 + 数据访问方法 (每个文件一个表)
Helper/helper.go                JWT 签发/解析、邮件发送、UUID/验证码生成、代码保存与校验
define/define.go                常量、共享类型 (ProblemBasic, TestCase)、环境变量引用
code/runner.go                 代码沙箱测试工具 (独立 main 包，非运行时组件)
docs/docs.go                   Swagger 自动生成文件 (勿手动编辑)
```

## 数据模型 (MySQL)

- **UserBasic** — 用户 (identity, name, password(MD5), mail, phone, pass_num, submit_num, is_admin)
- **ProblemBasic** — 题目 (identity, title, content, max_mem, max_runtime, submit_num, pass_num)
- **CategoryBasic** — 分类 (identity, name, parent_id)，树形结构
- **ProblemCategory** — 题目-分类关联表 (problem_id, category_id)
- **TestCase** — 测试用例 (identity, problem_identity, input, output)
- **SubmitBasic** — 提交记录 (identity, problem_identity, user_identity, path, status)

所有模型嵌入 `gorm.Model` (含 ID, CreatedAt, UpdatedAt, DeletedAt 软删除)。identity 字段用 UUID v4。

## 代码判题流程 (service/code.go)

1. 用户提交的 Go 代码保存到 `code/<uuid>/main.go`
2. `CheckGoCodeValid` 校验 import 白名单 (`bytes, fmt, math, sort, strings`)
3. 对每个测试用例并发执行 `go run <path>`，通过 stdin 传入 input
4. 对比 stdout 与预期 output；检查内存增量是否超限
5. 用 channel + select 监听 AC/WA/TLE/MLE/CE/非法代码 6 种状态
6. 超时控制通过 `time.After(MaxRuntime)` 实现（非硬件隔离）
7. 结果写入 submit_basic，并原子更新 user_basic 和 problem_basic 的计数

## API 约定

- 所有响应 HTTP 200，业务状态由 `code` 字段区分：`200` 成功，`-1` 失败
- 分页默认为 page=1, size=20
- 公共接口无需认证；`/user/*` 需用户 JWT (IsAdmin=0)；`/admin/*` 需管理员 JWT (IsAdmin=1)
- 前端通过 Vite 代理将 API 请求转发到 Go 后端 (:3000 → :8080)

## 注意事项

- DB 和 Redis 连接凭据硬编码在 `Models/init.go` 中，SMTP 凭据在 `Helper/helper.go` 的 `SendEmail` 中
- 密码使用 MD5 哈希，安全性较低
- `code/` 目录下的 UUID 子目录是用户提交代码的运行时产物，已加入 .gitignore
- 前端为单个 Vue SFC (`OnlineJudge.vue`)，所有模板/逻辑/样式集中在一个文件中
- 后端未使用 `.env` 文件，环境变量 `MysqlDNS` 和 `MailPassword` 在 `define/define.go` 中声明但实际未使用（凭据硬编码）
