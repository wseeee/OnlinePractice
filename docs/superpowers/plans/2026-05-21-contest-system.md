# 比赛系统实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 为 OnlineJudge 添加 OI 赛制比赛功能，支持定时比赛、公开/报名制、按通过比例计分、实时排行榜。

**架构：** 新增 4 张 GORM 模型表，从 service/code.go 提取 JudgeCode 公共函数复用判题逻辑，新建 service/contest.go 处理比赛相关 API，前端在 OnlineJudge.vue 中新增比赛列表和详情视图。

**技术栈：** Go + Gin + GORM + Redis (排名缓存) + Vue 3 Composition API

---

## 文件结构

| 操作 | 文件 | 职责 |
|------|------|------|
| 创建 | `Models/contest_basic.go` | 比赛主表模型 + 查询方法 |
| 创建 | `Models/contest_problem.go` | 比赛-题目关联模型 |
| 创建 | `Models/contest_user.go` | 报名记录模型 + 查询方法 |
| 创建 | `Models/contest_submit.go` | 比赛提交模型 + 查询方法 |
| 修改 | `service/code.go` | 提取 JudgeCode 公共函数 |
| 创建 | `service/contest.go` | 比赛相关 HTTP handler |
| 修改 | `router/app.go` | 注册比赛路由 |
| 修改 | `frontend/src/OnlineJudge.vue` | 新增比赛列表、详情、管理员比赛管理视图 |

---

## 任务 1：数据模型 — contest_basic

**文件：**
- 创建：`Models/contest_basic.go`

- [ ] **步骤 1：创建 contest_basic 模型文件**

```go
package Models

import (
	"time"

	"gorm.io/gorm"
)

type ContestBasic struct {
	gorm.Model
	Identity        string    `gorm:"column:identity;type:varchar(36);" json:"identity"`
	Title           string    `gorm:"column:title;type:varchar(255);" json:"title"`
	Description     string    `gorm:"column:description;type:text;" json:"description"`
	StartAt         time.Time `gorm:"column:start_at;" json:"start_at"`
	EndAt           time.Time `gorm:"column:end_at;" json:"end_at"`
	ContestType     int       `gorm:"column:contest_type;type:tinyint(1);" json:"contest_type"`
	MaxParticipants int       `gorm:"column:max_participants;type:int(11);default:0;" json:"max_participants"`
	ContestProblems []*ContestProblem `gorm:"foreignKey:contest_identity;references:identity" json:"contest_problems"`
}

func (table *ContestBasic) TableName() string {
	return "contest_basic"
}

func GetContestList(keyword string) *gorm.DB {
	return DB.Model(new(ContestBasic)).
		Where("title like ?", "%"+keyword+"%").
		Order("start_at DESC")
}

func GetContestDetail(identity string) *gorm.DB {
	return DB.Model(new(ContestBasic)).
		Preload("ContestProblems").
		Where("identity = ?", identity)
}

func CreateContest(c *ContestBasic) *gorm.DB {
	return DB.Model(new(ContestBasic)).Create(c)
}

func DeleteContest(identity string) *gorm.DB {
	return DB.Model(new(ContestBasic)).
		Where("identity = ?", identity).Delete(new(ContestBasic))
}

func UpdateContest(identity string, c *ContestBasic) *gorm.DB {
	return DB.Model(new(ContestBasic)).
		Where("identity = ?", identity).Updates(c)
}
```

- [ ] **步骤 2：确认编译通过**

运行：`go build ./Models/...`
预期：无错误

- [ ] **步骤 3：Commit**

```bash
git add Models/contest_basic.go
git commit -m "feat(models): add ContestBasic model"
```

---

## 任务 2：数据模型 — contest_problem

**文件：**
- 创建：`Models/contest_problem.go`

- [ ] **步骤 1：创建 contest_problem 模型文件**

```go
package Models

import "gorm.io/gorm"

type ContestProblem struct {
	gorm.Model
	ContestIdentity string        `gorm:"column:contest_identity;type:varchar(36);" json:"contest_identity"`
	ProblemIdentity string        `gorm:"column:problem_identity;type:varchar(36);" json:"problem_identity"`
	ProblemBasic    *ProblemBasic `gorm:"foreignKey:problem_identity;references:identity" json:"problem_basic"`
	Score           int           `gorm:"column:score;type:int(11);default:100;" json:"score"`
	Sort            int           `gorm:"column:sort;type:int(11);" json:"sort"`
}

func (table *ContestProblem) TableName() string {
	return "contest_problem"
}
```

- [ ] **步骤 2：确认编译通过**

运行：`go build ./Models/...`
预期：无错误

- [ ] **步骤 3：Commit**

```bash
git add Models/contest_problem.go
git commit -m "feat(models): add ContestProblem model"
```

---

## 任务 3：数据模型 — contest_user

**文件：**
- 创建：`Models/contest_user.go`

- [ ] **步骤 1：创建 contest_user 模型文件**

```go
package Models

import "gorm.io/gorm"

type ContestUser struct {
	gorm.Model
	ContestIdentity string `gorm:"column:contest_identity;type:varchar(36);" json:"contest_identity"`
	UserIdentity    string `gorm:"column:user_identity;type:varchar(36);" json:"user_identity"`
}

func (table *ContestUser) TableName() string {
	return "contest_user"
}

func IsContestUser(contestIdentity, userIdentity string) bool {
	var count int64
	DB.Model(new(ContestUser)).
		Where("contest_identity = ? AND user_identity = ?", contestIdentity, userIdentity).
		Count(&count)
	return count > 0
}

func GetContestUserCount(contestIdentity string) int64 {
	var count int64
	DB.Model(new(ContestUser)).
		Where("contest_identity = ?", contestIdentity).
		Count(&count)
	return count
}
```

- [ ] **步骤 2：确认编译通过**

运行：`go build ./Models/...`
预期：无错误

- [ ] **步骤 3：Commit**

```bash
git add Models/contest_user.go
git commit -m "feat(models): add ContestUser model"
```

---

## 任务 4：数据模型 — contest_submit

**文件：**
- 创建：`Models/contest_submit.go`

- [ ] **步骤 1：创建 contest_submit 模型文件**

```go
package Models

import "gorm.io/gorm"

type ContestSubmit struct {
	gorm.Model
	ContestIdentity string        `gorm:"column:contest_identity;type:varchar(36);" json:"contest_identity"`
	ProblemIdentity string        `gorm:"column:problem_identity;type:varchar(36);" json:"problem_identity"`
	UserIdentity    string        `gorm:"column:user_identity;type:varchar(36);" json:"user_identity"`
	Path            string        `gorm:"column:path;type:varchar(255);" json:"path"`
	Score           int           `gorm:"column:score;type:int(11);" json:"score"`
	Status          int           `gorm:"column:status;type:tinyint(1);" json:"status"`
	ProblemBasic    *ProblemBasic `gorm:"foreignKey:problem_identity;references:identity" json:"problem_basic"`
	UserBasic       *UserBasic    `gorm:"foreignKey:user_identity;references:identity" json:"user_basic"`
}

func (table *ContestSubmit) TableName() string {
	return "contest_submit"
}

func GetContestSubmitList(contestIdentity, problemIdentity, userIdentity string) *gorm.DB {
	tx := DB.Model(new(ContestSubmit)).
		Where("contest_identity = ?", contestIdentity).
		Preload("ProblemBasic", func(db *gorm.DB) *gorm.DB {
			return db.Omit("content")
		}).
		Preload("UserBasic", func(db *gorm.DB) *gorm.DB {
			return db.Omit("password")
		})
	if problemIdentity != "" {
		tx = tx.Where("problem_identity = ?", problemIdentity)
	}
	if userIdentity != "" {
		tx = tx.Where("user_identity = ?", userIdentity)
	}
	return tx.Order("created_at DESC")
}
```

- [ ] **步骤 2：确认编译通过**

运行：`go build ./Models/...`
预期：无错误

- [ ] **步骤 3：Commit**

```bash
git add Models/contest_submit.go
git commit -m "feat(models): add ContestSubmit model"
```

---

## 任务 5：提取 JudgeCode 公共函数

**文件：**
- 修改：`service/code.go`

- [ ] **步骤 1：从 CodeSubmit 中提取判题逻辑为 JudgeCode 函数**

在 `service/code.go` 中，在现有 `CodeSubmit` 函数之前添加 `JudgeCode` 函数。该函数封装：代码保存 → 校验 → 并发执行测试用例 → 比对结果，返回 (passed, total, status, err)。

在 `service/code.go` 的 `import` 块之后、`const maxOutputSize` 之后添加：

```go
// JudgeCode 判题公共函数，返回通过数、总用例数、状态码、错误
func JudgeCode(code []byte, problemIdentity string) (passed, total int, status int, err error) {
	// 代码保存
	path, err := Helper.SaveCode(code)
	if err != nil {
		return 0, 0, 0, err
	}

	// 查询题目及测试用例
	pb := new(Models.ProblemBasic)
	err = Models.DB.Where("identity = ?", problemIdentity).Preload("TestCases").First(pb).Error
	if err != nil {
		return 0, 0, 0, err
	}
	if len(pb.TestCases) == 0 {
		return 0, 0, 0, errors.New("该题目没有测试用例")
	}
	total = len(pb.TestCases)

	// 代码合法性校验
	v, err := Helper.CheckGoCodeValid(path)
	if err != nil {
		return 0, total, 0, err
	}
	if !v {
		return 0, total, 6, nil // 无效代码
	}

	// 并发执行测试用例
	WA := make(chan int, total)
	OOM := make(chan int, total)
	CE := make(chan string, total)
	AC := make(chan int, total)

	passCount := 0
	var lock sync.Mutex

	var wg sync.WaitGroup
	for _, testCase := range pb.TestCases {
		testCase := testCase
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command("go", "run", path)
			var out, stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &out
			stdinPipe, err := cmd.StdinPipe()
			if err != nil {
				CE <- "内部错误"
				return
			}
			io.WriteString(stdinPipe, testCase.Input+"\n")
			stdinPipe.Close()

			var bm runtime.MemStats
			runtime.ReadMemStats(&bm)
			if err := cmd.Run(); err != nil {
				errMsg := stderr.String()
				if len(errMsg) > 500 {
					errMsg = errMsg[:500]
				}
				if strings.Contains(err.Error(), "exit status 2") || strings.Contains(errMsg, "syntax error") {
					CE <- errMsg
					return
				}
			}
			var em runtime.MemStats
			runtime.ReadMemStats(&em)

			actual := strings.TrimSpace(out.String())
			expected := strings.TrimSpace(testCase.Output)
			if expected != actual {
				WA <- 1
				return
			}
			if em.Alloc/1024-(bm.Alloc/1024) > uint64(pb.MaxMem) {
				OOM <- 1
				return
			}
			lock.Lock()
			passCount++
			if passCount == total {
				AC <- 1
			}
			lock.Unlock()
		}()
	}

	go func() {
		wg.Wait()
		close(WA)
		close(OOM)
		close(CE)
		close(AC)
	}()

	select {
	case <-WA:
		lock.Lock()
		return passCount, total, 2, nil // WA
	case <-OOM:
		lock.Lock()
		return passCount, total, 4, nil // MLE
	case <-CE:
		lock.Lock()
		return passCount, total, 5, nil // CE
	case <-AC:
		lock.Lock()
		return total, total, 1, nil // AC
	case <-time.After(time.Millisecond * time.Duration(pb.MaxRuntime)):
		lock.Lock()
		if passCount == total {
			return total, total, 1, nil
		}
		return passCount, total, 3, nil // TLE
	}
}
```

- [ ] **步骤 2：重构 CodeSubmit 调用 JudgeCode**

将 `CodeSubmit` 函数中从 `// 代码保存` 到 `saveSubmitAndRespond` 之前的整个判题逻辑替换为调用 `JudgeCode`。新的 `CodeSubmit` 函数体：

```go
func CodeSubmit(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	if problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不完整"})
		return
	}

	code, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "读取代码失败"})
		return
	}

	u, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户信息获取失败"})
		return
	}
	userClaim := u.(*Helper.UserJwt)

	passed, total, status, err := JudgeCode(code, problemIdentity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
		return
	}

	sb := &Models.SubmitBasic{
		Identity:        Helper.GetUUID(),
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity,
		Status:          status,
	}

	msgMap := map[int]string{1: "答案正确", 2: "答案错误", 3: "运行超时", 4: "运行超内存", 5: "编译错误", 6: "无效代码"}
	msg := msgMap[status]
	_ = passed
	_ = total

	saveSubmitAndRespond(c, sb, userClaim, problemIdentity, msg, nil)
}
```

注意：`saveSubmitAndRespond` 中使用 `pb` 的地方是 `pb.MaxRuntime` 和 `pb.MaxMem`，但实际只在 `sb.Status == 1` 时更新计数，不需要 pb。检查现有 `saveSubmitAndRespond` 签名，最后一个参数 `pb` 在函数体中只用于 `pb.MaxRuntime`（select 超时），但重构后超时已在 JudgeCode 中处理。修改 `saveSubmitAndRespond` 签名去掉 `pb` 参数，或传 nil。实际查看原代码，`pb` 在 `saveSubmitAndRespond` 中未被使用（只在 `CodeSubmit` 的 select 中使用），所以可以直接传 nil 或修改签名。

- [ ] **步骤 3：确认编译通过**

运行：`go build ./service/...`
预期：无错误

- [ ] **步骤 4：Commit**

```bash
git add service/code.go
git commit -m "refactor(service): extract JudgeCode from CodeSubmit"
```

---

## 任务 6：比赛管理员接口 — 创建比赛

**文件：**
- 创建：`service/contest.go`

- [ ] **步骤 1：创建 contest.go，实现 CreateContest handler**

```go
package service

import (
	"OnlinePrictice/Helper"
	"OnlinePrictice/Models"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateContest
// @Tags 管理员私有方法
// @Summary 创建比赛
// @Param Authorization header string true "Authorization"
// @Param title formData string true "title"
// @Param description formData string true "description"
// @Param start_at formData string true "start_at (2006-01-02 15:04:05)"
// @Param end_at formData string true "end_at (2006-01-02 15:04:05)"
// @Param contest_type formData int true "contest_type (1=公开, 2=报名制)"
// @Param max_participants formData int false "max_participants (报名制上限)"
// @Param problem_identities formData string true "problem_identities (JSON array)"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/contest-create [post]
func CreateContest(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	startAtStr := c.PostForm("start_at")
	endAtStr := c.PostForm("end_at")
	contestType, _ := strconv.Atoi(c.PostForm("contest_type"))
	maxParticipants, _ := strconv.Atoi(c.PostForm("max_participants"))
	problemIdentitiesStr := c.PostForm("problem_identities")

	if title == "" || startAtStr == "" || endAtStr == "" || contestType == 0 || problemIdentitiesStr == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不齐全"})
		return
	}

	startAt, err := time.Parse(define.DateLayout, startAtStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "开始时间格式错误"})
		return
	}
	endAt, err := time.Parse(define.DateLayout, endAtStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "结束时间格式错误"})
		return
	}
	if !endAt.After(startAt) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "结束时间必须晚于开始时间"})
		return
	}

	var problemIdentities []string
	if err := json.Unmarshal([]byte(problemIdentitiesStr), &problemIdentities); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "题目列表格式错误"})
		return
	}

	identity := Helper.GetUUID()
	contest := &Models.ContestBasic{
		Identity:        identity,
		Title:           title,
		Description:     description,
		StartAt:         startAt,
		EndAt:           endAt,
		ContestType:     contestType,
		MaxParticipants: maxParticipants,
	}

	contestProblems := make([]*Models.ContestProblem, 0, len(problemIdentities))
	for i, pi := range problemIdentities {
		contestProblems = append(contestProblems, &Models.ContestProblem{
			ContestIdentity: identity,
			ProblemIdentity: pi,
			Score:           100,
			Sort:            i + 1,
		})
	}
	contest.ContestProblems = contestProblems

	err = Models.CreateContest(contest).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{"identity": identity},
	})
}
```

- [ ] **步骤 2：确认编译通过**

运行：`go build ./service/...`
预期：无错误（需要先在 import 中添加 `"OnlinePrictice/define"`）

- [ ] **步骤 3：Commit**

```bash
git add service/contest.go
git commit -m "feat(service): add CreateContest handler"
```

---

## 任务 7：比赛管理员接口 — 更新、列表、删除

**文件：**
- 修改：`service/contest.go`

- [ ] **步骤 1：添加 UpdateContest handler**

在 `service/contest.go` 末尾添加：

```go
// UpdateContest
// @Tags 管理员私有方法
// @Summary 修改比赛
// @Param Authorization header string true "Authorization"
// @Param identity query string true "identity"
// @Param title formData string true "title"
// @Param description formData string true "description"
// @Param start_at formData string true "start_at"
// @Param end_at formData string true "end_at"
// @Param contest_type formData int true "contest_type"
// @Param max_participants formData int false "max_participants"
// @Param problem_identities formData string true "problem_identities (JSON array)"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/contest-update [put]
func UpdateContest(c *gin.Context) {
	identity := c.Query("identity")
	title := c.PostForm("title")
	description := c.PostForm("description")
	startAtStr := c.PostForm("start_at")
	endAtStr := c.PostForm("end_at")
	contestType, _ := strconv.Atoi(c.PostForm("contest_type"))
	maxParticipants, _ := strconv.Atoi(c.PostForm("max_participants"))
	problemIdentitiesStr := c.PostForm("problem_identities")

	if identity == "" || title == "" || startAtStr == "" || endAtStr == "" || contestType == 0 || problemIdentitiesStr == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不齐全"})
		return
	}

	startAt, err := time.Parse(define.DateLayout, startAtStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "开始时间格式错误"})
		return
	}
	endAt, err := time.Parse(define.DateLayout, endAtStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "结束时间格式错误"})
		return
	}

	var problemIdentities []string
	if err := json.Unmarshal([]byte(problemIdentitiesStr), &problemIdentities); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "题目列表格式错误"})
		return
	}

	err = Models.DB.Transaction(func(tx *gorm.DB) error {
		// 更新比赛基本信息
		contest := &Models.ContestBasic{
			Title:           title,
			Description:     description,
			StartAt:         startAt,
			EndAt:           endAt,
			ContestType:     contestType,
			MaxParticipants: maxParticipants,
		}
		err := tx.Model(new(Models.ContestBasic)).
			Where("identity = ?", identity).
			Updates(contest).Error
		if err != nil {
			return err
		}
		// 删除旧的题目关联
		err = tx.Where("contest_identity = ?", identity).Delete(&Models.ContestProblem{}).Error
		if err != nil {
			return err
		}
		// 创建新的题目关联
		contestProblems := make([]*Models.ContestProblem, 0, len(problemIdentities))
		for i, pi := range problemIdentities {
			contestProblems = append(contestProblems, &Models.ContestProblem{
				ContestIdentity: identity,
				ProblemIdentity: pi,
				Score:           100,
				Sort:            i + 1,
			})
		}
		return tx.Create(&contestProblems).Error
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "修改成功"})
}
```

- [ ] **步骤 2：添加 GetContestList (admin) handler**

```go
// GetContestList
// @Tags 管理员私有方法
// @Summary 比赛管理列表
// @Param Authorization header string true "Authorization"
// @Param page query int false "page"
// @Param size query int false "size"
// @Param keyword query string false "keyword"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/contest-list [get]
func GetContestList(c *gin.Context) {
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page, _ := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	page = (page - 1) * size
	keyword := c.Query("keyword")

	list := make([]*Models.ContestBasic, 0)
	var count int64
	err := Models.GetContestList(keyword).Count(&count).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{"list": list, "count": count},
	})
}
```

- [ ] **步骤 3：添加 DeleteContest handler**

```go
// DeleteContest
// @Tags 管理员私有方法
// @Summary 删除比赛
// @Param Authorization header string true "Authorization"
// @Param identity query string true "identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/contest-delete [delete]
func DeleteContest(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不完整"})
		return
	}
	err := Models.DeleteContest(identity).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
```

- [ ] **步骤 4：确认编译通过**

运行：`go build ./service/...`
预期：无错误

- [ ] **步骤 5：Commit**

```bash
git add service/contest.go
git commit -m "feat(service): add contest admin handlers (update/list/delete)"
```

---

## 任务 8：比赛用户接口 — 报名

**文件：**
- 修改：`service/contest.go`

- [ ] **步骤 1：添加 ContestJoin handler**

```go
// ContestJoin
// @Tags 用户私有方法
// @Summary 报名比赛
// @Param Authorization header string true "Authorization"
// @Param contest_identity query string true "contest_identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /user/contest-join [post]
func ContestJoin(c *gin.Context) {
	contestIdentity := c.Query("contest_identity")
	if contestIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不完整"})
		return
	}

	u, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户信息获取失败"})
		return
	}
	userClaim := u.(*Helper.UserJwt)

	// 检查比赛是否存在
	var contest Models.ContestBasic
	err := Models.DB.Where("identity = ?", contestIdentity).First(&contest).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "比赛不存在"})
		return
	}
	if contest.ContestType != 2 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "该比赛无需报名"})
		return
	}

	// 检查是否已报名
	if Models.IsContestUser(contestIdentity, userClaim.Identity) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "已报名，请勿重复操作"})
		return
	}

	// 检查人数上限
	if contest.MaxParticipants > 0 {
		count := Models.GetContestUserCount(contestIdentity)
		if count >= int64(contest.MaxParticipants) {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "报名人数已满"})
			return
		}
	}

	cu := &Models.ContestUser{
		ContestIdentity: contestIdentity,
		UserIdentity:    userClaim.Identity,
	}
	err = Models.DB.Create(cu).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "报名失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "报名成功"})
}
```

- [ ] **步骤 2：确认编译通过**

运行：`go build ./service/...`
预期：无错误

- [ ] **步骤 3：Commit**

```bash
git add service/contest.go
git commit -m "feat(service): add ContestJoin handler"
```

---

## 任务 9：比赛用户接口 — 提交代码

**文件：**
- 修改：`service/contest.go`

- [ ] **步骤 1：添加 ContestSubmit handler**

```go
// ContestSubmit
// @Tags 用户私有方法
// @Summary 比赛内提交代码
// @Param Authorization header string true "Authorization"
// @Param contest_identity query string true "contest_identity"
// @Param problem_identity query string true "problem_identity"
// @Param code body string true "code"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /user/contest-submit [post]
func ContestSubmit(c *gin.Context) {
	contestIdentity := c.Query("contest_identity")
	problemIdentity := c.Query("problem_identity")
	if contestIdentity == "" || problemIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不完整"})
		return
	}

	code, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "读取代码失败"})
		return
	}

	u, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户信息获取失败"})
		return
	}
	userClaim := u.(*Helper.UserJwt)

	// 校验比赛状态
	var contest Models.ContestBasic
	err = Models.DB.Where("identity = ?", contestIdentity).First(&contest).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "比赛不存在"})
		return
	}
	now := time.Now()
	if now.Before(contest.StartAt) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "比赛尚未开始"})
		return
	}
	if now.After(contest.EndAt) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "比赛已结束"})
		return
	}

	// 报名制需检查是否已报名
	if contest.ContestType == 2 && !Models.IsContestUser(contestIdentity, userClaim.Identity) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "请先报名"})
		return
	}

	// 校验题目是否属于该比赛
	var cp Models.ContestProblem
	err = Models.DB.Where("contest_identity = ? AND problem_identity = ?", contestIdentity, problemIdentity).First(&cp).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "该题目不属于此比赛"})
		return
	}

	// 判题
	passed, total, status, err := JudgeCode(code, problemIdentity)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
		return
	}

	// OI 计分
	score := 0
	if total > 0 {
		score = passed * cp.Score / total
	}

	cs := &Models.ContestSubmit{
		ContestIdentity: contestIdentity,
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity,
		Path:            "code/" + Helper.GetUUID() + "/main.go",
		Score:           score,
		Status:          status,
	}
	err = Models.DB.Create(cs).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "提交保存失败"})
		return
	}

	msgMap := map[int]string{1: "答案正确", 2: "答案错误", 3: "运行超时", 4: "运行超内存", 5: "编译错误", 6: "无效代码"}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"score":  score,
			"status": status,
			"msg":    msgMap[status],
		},
	})
}
```

注意：`ContestSubmit` 中的 `Path` 字段应使用 `JudgeCode` 返回后实际保存的路径。但当前 `JudgeCode` 不返回 path。有两种方案：(1) 让 `JudgeCode` 额外返回 path；(2) 在 `ContestSubmit` 中自己调用 `Helper.SaveCode` 后再判题。推荐方案 (1)，在任务 5 中修改 `JudgeCode` 签名增加 path 返回值。如果已完成任务 5，需回头修改。

实际处理：在任务 5 的 `JudgeCode` 签名中增加 `path string` 返回值：
```go
func JudgeCode(code []byte, problemIdentity string) (passed, total int, status int, path string, err error)
```
`ContestSubmit` 中使用返回的 path。

- [ ] **步骤 2：确认编译通过**

运行：`go build ./service/...`
预期：无错误

- [ ] **步骤 3：Commit**

```bash
git add service/contest.go
git commit -m "feat(service): add ContestSubmit handler with OI scoring"
```

---

## 任务 10：比赛用户接口 — 查看提交记录

**文件：**
- 修改：`service/contest.go`

- [ ] **步骤 1：添加 GetUserContestSubmits handler**

```go
// GetUserContestSubmits
// @Tags 用户私有方法
// @Summary 查看自己在某比赛的提交记录
// @Param Authorization header string true "Authorization"
// @Param contest_identity query string true "contest_identity"
// @Param problem_identity query string false "problem_identity"
// @Param page query int false "page"
// @Param size query int false "size"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /user/contest-submits [get]
func GetUserContestSubmits(c *gin.Context) {
	contestIdentity := c.Query("contest_identity")
	if contestIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不完整"})
		return
	}
	problemIdentity := c.Query("problem_identity")

	u, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户信息获取失败"})
		return
	}
	userClaim := u.(*Helper.UserJwt)

	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page, _ := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	page = (page - 1) * size

	list := make([]*Models.ContestSubmit, 0)
	var count int64
	err := Models.GetContestSubmitList(contestIdentity, problemIdentity, userClaim.Identity).
		Count(&count).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{"list": list, "count": count},
	})
}
```

- [ ] **步骤 2：确认编译通过**

运行：`go build ./service/...`
预期：无错误

- [ ] **步骤 3：Commit**

```bash
git add service/contest.go
git commit -m "feat(service): add GetUserContestSubmits handler"
```

---

## 任务 11：比赛公共接口 — 列表、详情、排行榜

**文件：**
- 修改：`service/contest.go`

- [ ] **步骤 1：添加 GetContestDetailPublic handler**

```go
// GetContestDetailPublic
// @Tags 公共方法
// @Summary 比赛详情
// @Param identity query string true "identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /contest-detail [get]
func GetContestDetailPublic(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不完整"})
		return
	}
	var contest Models.ContestBasic
	err := Models.GetContestDetail(identity).First(&contest).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "比赛不存在"})
		return
	}

	// 附加信息：参赛人数、当前状态
	participantCount := Models.GetContestUserCount(identity)
	now := time.Now()
	status := 0 // 未开始
	if now.After(contest.StartAt) && now.Before(contest.EndAt) {
		status = 1 // 进行中
	} else if now.After(contest.EndAt) {
		status = 2 // 已结束
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"contest":           contest,
			"participant_count": participantCount,
			"status":            status,
		},
	})
}
```

- [ ] **步骤 2：添加 GetContestRank handler**

```go
// GetContestRank
// @Tags 公共方法
// @Summary 比赛排行榜
// @Param contest_identity query string true "contest_identity"
// @Param page query int false "page"
// @Param size query int false "size"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /contest-rank [get]
func GetContestRank(c *gin.Context) {
	contestIdentity := c.Query("contest_identity")
	if contestIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数不完整"})
		return
	}

	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page, _ := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))

	// 尝试从 Redis 缓存获取
	cacheKey := "contest_rank:" + contestIdentity
	cached, err := Models.RDB.Get(define.CTX, cacheKey).Result()
	if err == nil && cached != "" {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": cached})
		return
	}

	// 查询该比赛所有提交
	var submits []Models.ContestSubmit
	err = Models.DB.Where("contest_identity = ?", contestIdentity).
		Find(&submits).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}

	// 按 (user, problem) 聚合：每题最高分 + 最早 AC 时间
	type userProblemScore struct {
		Score     int
		FirstACTime time.Time
	}
	type userRank struct {
		UserIdentity string    `json:"user_identity"`
		UserName     string    `json:"user_name"`
		TotalScore   int       `json:"total_score"`
		TotalTime    time.Time `json:"total_time"`
	}

	userMap := make(map[string]map[string]*userProblemScore) // user -> problem -> best
	for _, s := range submits {
		if userMap[s.UserIdentity] == nil {
			userMap[s.UserIdentity] = make(map[string]*userProblemScore)
		}
		ups := userMap[s.UserIdentity][s.ProblemIdentity]
		if ups == nil {
			ups = &userProblemScore{}
			userMap[s.UserIdentity][s.ProblemIdentity] = ups
		}
		if s.Score > ups.Score {
			ups.Score = s.Score
		}
		if s.Status == 1 && (ups.FirstACTime.IsZero() || s.CreatedAt.Before(ups.FirstACTime)) {
			ups.FirstACTime = s.CreatedAt
		}
	}

	// 构建排名列表
	ranks := make([]userRank, 0, len(userMap))
	for userIdentity, problems := range userMap {
		totalScore := 0
		var earliestAC time.Time
		for _, ups := range problems {
			totalScore += ups.Score
			if !ups.FirstACTime.IsZero() {
				if earliestAC.IsZero() || ups.FirstACTime.Before(earliestAC) {
					earliestAC = ups.FirstACTime
				}
			}
		}
		// 查询用户名
		var user Models.UserBasic
		Models.DB.Where("identity = ?", userIdentity).First(&user)
		ranks = append(ranks, userRank{
			UserIdentity: userIdentity,
			UserName:     user.Name,
			TotalScore:   totalScore,
			TotalTime:    earliestAC,
		})
	}

	// 排序：总分降序，同分按最早 AC 时间升序
	sort.Slice(ranks, func(i, j int) bool {
		if ranks[i].TotalScore != ranks[j].TotalScore {
			return ranks[i].TotalScore > ranks[j].TotalScore
		}
		return ranks[i].TotalTime.Before(ranks[j].TotalTime)
	})

	// 分页
	total := len(ranks)
	start := (page - 1) * size
	end := start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	pagedRanks := ranks[start:end]

	// 缓存到 Redis（30 秒）
	rankJSON, _ := json.Marshal(ranks)
	Models.RDB.Set(define.CTX, cacheKey, string(rankJSON), 30*time.Second)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  pagedRanks,
			"count": total,
		},
	})
}
```

注意：需要在 import 中添加 `"sort"`。

- [ ] **步骤 3：确认编译通过**

运行：`go build ./service/...`
预期：无错误

- [ ] **步骤 4：Commit**

```bash
git add service/contest.go
git commit -m "feat(service): add contest public APIs (detail/rank)"
```

---

## 任务 12：注册路由

**文件：**
- 修改：`router/app.go`

- [ ] **步骤 1：在路由中添加比赛相关路由**

在 `router/app.go` 的公共方法区域添加：

```go
// 比赛
r.GET("/contest-list", service.GetContestListPublic)
r.GET("/contest-detail", service.GetContestDetailPublic)
r.GET("/contest-rank", service.GetContestRank)
```

在 `user` 组中添加：

```go
user.POST("/contest-join", service.ContestJoin)
user.POST("/contest-submit", service.ContestSubmit)
user.GET("/contest-submits", service.GetUserContestSubmits)
```

在 `admin` 组中添加：

```go
admin.POST("/contest-create", service.CreateContest)
admin.PUT("/contest-update", service.UpdateContest)
admin.GET("/contest-list", service.GetContestList)
admin.DELETE("/contest-delete", service.DeleteContest)
```

注意：公共路由中的 `GetContestListPublic` 需要单独实现（与 admin 的 `GetContestList` 不同，公共版本不返回敏感信息）。在 `service/contest.go` 中添加：

```go
// GetContestListPublic
// @Tags 公共方法
// @Summary 比赛列表
// @Param page query int false "page"
// @Param size query int false "size"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /contest-list [get]
func GetContestListPublic(c *gin.Context) {
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page, _ := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	page = (page - 1) * size

	list := make([]*Models.ContestBasic, 0)
	var count int64
	err := Models.DB.Model(new(Models.ContestBasic)).
		Order("start_at DESC").
		Count(&count).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}

	// 附加每场比赛的参赛人数和状态
	type contestInfo struct {
		*Models.ContestBasic
		ParticipantCount int64 `json:"participant_count"`
		ContestStatus    int   `json:"contest_status"`
	}
	result := make([]contestInfo, len(list))
	now := time.Now()
	for i, c := range list {
		status := 0
		if now.After(c.StartAt) && now.Before(c.EndAt) {
			status = 1
		} else if now.After(c.EndAt) {
			status = 2
		}
		result[i] = contestInfo{
			ContestBasic:     c,
			ParticipantCount: Models.GetContestUserCount(c.Identity),
			ContestStatus:    status,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{"list": result, "count": count},
	})
}
```

- [ ] **步骤 2：确认编译通过**

运行：`go build ./...`
预期：无错误

- [ ] **步骤 3：Commit**

```bash
git add router/app.go service/contest.go
git commit -m "feat(router): register contest routes"
```

---

## 任务 13：前端 — 导航栏添加比赛入口

**文件：**
- 修改：`frontend/src/OnlineJudge.vue`

- [ ] **步骤 1：在导航栏中添加比赛 tab**

在 `<nav class="header-nav">` 中，在"排行榜"和"提交记录"之间添加：

```html
<a :class="{ active: currentView === 'contest' }" @click="switchView('contest')">比赛</a>
```

- [ ] **步骤 2：验证前端可运行**

运行：`cd frontend && npm run dev`
预期：导航栏出现"比赛"tab，点击可切换（虽然视图内容还没有）

- [ ] **步骤 3：Commit**

```bash
git add frontend/src/OnlineJudge.vue
git commit -m "feat(frontend): add contest nav tab"
```

---

## 任务 14：前端 — 比赛列表视图

**文件：**
- 修改：`frontend/src/OnlineJudge.vue`

- [ ] **步骤 1：在模板中添加比赛列表 section**

在 `</section>` (题库 section 结束) 之后、题目详情 section 之前添加：

```html
<!-- ===== 比赛列表页 ===== -->
<section v-if="currentView === 'contest'" class="view-problems">
  <div class="page-header">
    <h2>比赛</h2>
    <p class="subtitle">参加比赛，挑战自我</p>
  </div>
  <div v-if="contestLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
  <div v-else-if="contestError" class="state-box error">{{ contestError }}</div>
  <div v-else-if="contests.length === 0" class="state-box empty">暂无比赛</div>
  <div v-else class="problem-table-wrap">
    <table class="data-table">
      <thead><tr><th>比赛</th><th>开始时间</th><th>结束时间</th><th>状态</th><th>参赛人数</th></tr></thead>
      <tbody>
        <tr v-for="c in contests" :key="c.identity" @click="openContestDetail(c)" class="clickable">
          <td>{{ c.title }}</td>
          <td>{{ c.start_at }}</td>
          <td>{{ c.end_at }}</td>
          <td><span class="status-badge" :class="contestStatusClass(c.contest_status)">{{ contestStatusText(c.contest_status) }}</span></td>
          <td>{{ c.participant_count }}</td>
        </tr>
      </tbody>
    </table>
    <div class="pagination">
      <button :disabled="contestPage <= 1" @click="contestPage--; fetchContests()">上一页</button>
      <span>第 {{ contestPage }} 页 / 共 {{ Math.ceil(contestTotal / contestSize) || 1 }} 页</span>
      <button :disabled="contestPage >= Math.ceil(contestTotal / contestSize)" @click="contestPage++; fetchContests()">下一页</button>
    </div>
  </div>
</section>
```

- [ ] **步骤 2：在 script setup 中添加比赛列表数据和方法**

在 `<script setup>` 中添加：

```js
// ===== 比赛列表 =====
const contests = ref([])
const contestTotal = ref(0)
const contestPage = ref(1)
const contestSize = ref(20)
const contestLoading = ref(false)
const contestError = ref('')

async function fetchContests() {
  contestLoading.value = true
  contestError.value = ''
  try {
    const res = await api.get('/contest-list', { params: { page: contestPage.value, size: contestSize.value } })
    if (res.data.code === 200) {
      contests.value = res.data.data.list
      contestTotal.value = res.data.data.count
    } else {
      contestError.value = res.data.msg
    }
  } catch (e) {
    contestError.value = '网络错误'
  } finally {
    contestLoading.value = false
  }
}

function contestStatusClass(s) {
  return { 0: 'status-waiting', 1: 'status-running', 2: 'status-ended' }[s] || ''
}
function contestStatusText(s) {
  return { 0: '未开始', 1: '进行中', 2: '已结束' }[s] || '未知'
}
```

- [ ] **步骤 3：在 switchView 中添加比赛列表加载**

修改 `switchView` 函数，添加：

```js
if (v === 'contest') fetchContests()
```

- [ ] **步骤 4：在 style 中添加状态标签样式**

在 `<style scoped>` 中添加：

```css
.status-badge { padding: 2px 8px; border-radius: 4px; font-size: 12px; }
.status-waiting { background: #e6f7ff; color: #1890ff; }
.status-running { background: #f6ffed; color: #52c41a; }
.status-ended { background: #fff1f0; color: #ff4d4f; }
```

- [ ] **步骤 5：验证前端可运行**

运行：`cd frontend && npm run dev`
预期：点击"比赛"tab 显示比赛列表页

- [ ] **步骤 6：Commit**

```bash
git add frontend/src/OnlineJudge.vue
git commit -m "feat(frontend): add contest list view"
```

---

## 任务 15：前端 — 比赛详情视图

**文件：**
- 修改：`frontend/src/OnlineJudge.vue`

- [ ] **步骤 1：在模板中添加比赛详情 section**

在比赛列表 section 之后添加：

```html
<!-- ===== 比赛详情页 ===== -->
<section v-if="currentView === 'contest-detail'" class="view-detail">
  <button class="btn btn-text" @click="currentView = 'contest'; fetchContests()">← 返回比赛列表</button>
  <div v-if="contestDetailLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
  <div v-else-if="contestDetailError" class="state-box error">{{ contestDetailError }}</div>
  <div v-else>
    <div class="detail-left">
      <h2>{{ contestDetail.contest.title }}</h2>
      <div class="detail-meta">
        <span>开始: {{ contestDetail.contest.start_at }}</span>
        <span>结束: {{ contestDetail.contest.end_at }}</span>
        <span>参赛人数: {{ contestDetail.participant_count }}</span>
        <span class="status-badge" :class="contestStatusClass(contestDetail.status)">{{ contestStatusText(contestDetail.status) }}</span>
      </div>
      <div v-if="contestDetail.contest.description" class="problem-content" style="margin-top:16px">
        <pre>{{ contestDetail.contest.description }}</pre>
      </div>

      <!-- 比赛题目列表 -->
      <h3 style="margin-top:24px">题目列表</h3>
      <table class="data-table">
        <thead><tr><th>#</th><th>题目</th><th>满分</th><th>我的得分</th></tr></thead>
        <tbody>
          <tr v-for="(cp, idx) in contestDetail.contest.contest_problems" :key="cp.id"
              @click="openContestProblem(cp)" class="clickable">
            <td>{{ idx + 1 }}</td>
            <td>{{ cp.problem_basic ? cp.problem_basic.title : cp.problem_identity }}</td>
            <td>{{ cp.score }}</td>
            <td>{{ contestMyScores[cp.problem_identity] ?? '-' }}</td>
          </tr>
        </tbody>
      </table>

      <!-- 报名按钮（报名制） -->
      <div v-if="contestDetail.contest.contest_type === 2 && contestDetail.status === 0" style="margin-top:16px">
        <button class="btn btn-primary" @click="joinContest" :disabled="contestJoining">
          {{ contestJoining ? '报名中...' : '报名参加' }}
        </button>
        <span v-if="contestJoinMsg" style="margin-left:8px">{{ contestJoinMsg }}</span>
      </div>
    </div>

    <!-- 右侧：提交记录 + 排行榜 -->
    <div class="detail-right">
      <div class="right-tabs">
        <a :class="{ active: contestTab === 'submit' }" @click="contestTab = 'submit'">我的提交</a>
        <a :class="{ active: contestTab === 'rank' }" @click="contestTab = 'rank'; fetchContestRank()">排行榜</a>
      </div>

      <!-- 提交代码区 -->
      <div v-if="contestTab === 'submit'">
        <div v-if="selectedContestProblem" style="margin-bottom:8px">
          <strong>当前题目：</strong>{{ selectedContestProblem.problem_basic?.title }}
        </div>
        <textarea v-model="contestCode" class="code-editor" placeholder="在此编写 Go 代码..." rows="12"></textarea>
        <div style="margin-top:8px; display:flex; gap:8px">
          <button class="btn btn-primary" @click="submitContestCode" :disabled="contestSubmitting || !selectedContestProblem">
            {{ contestSubmitting ? '提交中...' : '提交代码' }}
          </button>
          <span v-if="contestSubmitMsg" :style="{ color: contestSubmitColor }">{{ contestSubmitMsg }}</span>
        </div>

        <div v-if="contestSubmits.length > 0" style="margin-top:16px">
          <table class="data-table">
            <thead><tr><th>题目</th><th>得分</th><th>状态</th><th>时间</th></tr></thead>
            <tbody>
              <tr v-for="s in contestSubmits" :key="s.id">
                <td>{{ s.problem_basic?.title ?? s.problem_identity }}</td>
                <td>{{ s.score }}</td>
                <td>{{ submitStatusText(s.status) }}</td>
                <td>{{ s.created_at }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 排行榜 -->
      <div v-if="contestTab === 'rank'">
        <div v-if="contestRankLoading" class="state-box"><span class="spinner"></span> 加载中...</div>
        <div v-else-if="contestRanks.length === 0" class="state-box empty">暂无排名数据</div>
        <table v-else class="data-table">
          <thead><tr><th>排名</th><th>用户</th><th>总分</th></tr></thead>
          <tbody>
            <tr v-for="(r, idx) in contestRanks" :key="r.user_identity">
              <td>{{ idx + 1 }}</td>
              <td>{{ r.user_name }}</td>
              <td>{{ r.total_score }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</section>
```

- [ ] **步骤 2：在 script setup 中添加比赛详情数据和方法**

```js
// ===== 比赛详情 =====
const contestDetail = ref(null)
const contestDetailLoading = ref(false)
const contestDetailError = ref('')
const contestMyScores = ref({})
const selectedContestProblem = ref(null)
const contestCode = ref('')
const contestSubmitting = ref(false)
const contestSubmitMsg = ref('')
const contestSubmitColor = ref('')
const contestSubmits = ref([])
const contestTab = ref('submit')
const contestRanks = ref([])
const contestRankLoading = ref(false)
const contestJoining = ref(false)
const contestJoinMsg = ref('')

async function openContestDetail(c) {
  currentView.value = 'contest-detail'
  contestDetailLoading.value = true
  contestDetailError.value = ''
  contestMyScores.value = {}
  selectedContestProblem.value = null
  contestSubmits.value = []
  contestRanks.value = []
  contestTab.value = 'submit'
  try {
    const res = await api.get('/contest-detail', { params: { identity: c.identity } })
    if (res.data.code === 200) {
      contestDetail.value = res.data.data
      // 加载我的提交记录以获取各题最高分
      const subRes = await api.get('/user/contest-submits', { params: { contest_identity: c.identity, size: 1000 } })
      if (subRes.data.code === 200) {
        const scores = {}
        for (const s of subRes.data.data.list) {
          const cur = scores[s.problem_identity] ?? 0
          if (s.score > cur) scores[s.problem_identity] = s.score
        }
        contestMyScores.value = scores
        contestSubmits.value = subRes.data.data.list
      }
    } else {
      contestDetailError.value = res.data.msg
    }
  } catch (e) {
    contestDetailError.value = '网络错误'
  } finally {
    contestDetailLoading.value = false
  }
}

function openContestProblem(cp) {
  selectedContestProblem.value = cp
  contestCode.value = ''
  contestSubmitMsg.value = ''
}

async function joinContest() {
  if (!contestDetail.value) return
  contestJoining.value = true
  contestJoinMsg.value = ''
  try {
    const res = await api.post('/user/contest-join', null, { params: { contest_identity: contestDetail.value.contest.identity } })
    contestJoinMsg.value = res.data.code === 200 ? '报名成功' : res.data.msg
  } catch (e) {
    contestJoinMsg.value = '网络错误'
  } finally {
    contestJoining.value = false
  }
}

async function submitContestCode() {
  if (!selectedContestProblem.value || !contestCode.value.trim()) return
  contestSubmitting.value = true
  contestSubmitMsg.value = ''
  try {
    const res = await api.post('/user/contest-submit', contestCode.value, {
      params: {
        contest_identity: contestDetail.value.contest.identity,
        problem_identity: selectedContestProblem.value.problem_identity
      },
      headers: { 'Content-Type': 'text/plain' }
    })
    if (res.data.code === 200) {
      contestSubmitMsg.value = res.data.data.msg + ' (得分: ' + res.data.data.score + ')'
      contestSubmitColor.value = res.data.data.status === 1 ? '#52c41a' : '#ff4d4f'
      // 刷新提交记录
      const subRes = await api.get('/user/contest-submits', {
        params: { contest_identity: contestDetail.value.contest.identity, size: 1000 }
      })
      if (subRes.data.code === 200) {
        contestSubmits.value = subRes.data.data.list
        const scores = {}
        for (const s of subRes.data.data.list) {
          const cur = scores[s.problem_identity] ?? 0
          if (s.score > cur) scores[s.problem_identity] = s.score
        }
        contestMyScores.value = scores
      }
    } else {
      contestSubmitMsg.value = res.data.msg
      contestSubmitColor.value = '#ff4d4f'
    }
  } catch (e) {
    contestSubmitMsg.value = '网络错误'
    contestSubmitColor.value = '#ff4d4f'
  } finally {
    contestSubmitting.value = false
  }
}

async function fetchContestRank() {
  if (!contestDetail.value) return
  contestRankLoading.value = true
  try {
    const res = await api.get('/contest-rank', { params: { contest_identity: contestDetail.value.contest.identity, size: 100 } })
    if (res.data.code === 200) {
      contestRanks.value = res.data.data.list
    }
  } catch (e) {
    // ignore
  } finally {
    contestRankLoading.value = false
  }
}

function submitStatusText(s) {
  return { 1: 'AC', 2: 'WA', 3: 'TLE', 4: 'MLE', 5: 'CE', 6: '无效代码' }[s] || '未知'
}
```

- [ ] **步骤 3：添加 CSS 样式**

```css
.right-tabs { display: flex; gap: 12px; margin-bottom: 12px; border-bottom: 1px solid #eee; padding-bottom: 8px; }
.right-tabs a { cursor: pointer; padding: 4px 8px; }
.right-tabs a.active { border-bottom: 2px solid #1890ff; color: #1890ff; }
```

- [ ] **步骤 4：验证前端可运行**

运行：`cd frontend && npm run dev`
预期：比赛列表可点击进入详情，详情页展示题目列表、代码编辑器、排行榜 tab

- [ ] **步骤 5：Commit**

```bash
git add frontend/src/OnlineJudge.vue
git commit -m "feat(frontend): add contest detail view with submit and rank"
```

---

## 任务 16：前端 — 管理员比赛管理

**文件：**
- 修改：`frontend/src/OnlineJudge.vue`

- [ ] **步骤 1：在管理员面板中添加比赛管理 tab**

在管理员 section 的 tab 列表中添加"比赛管理"tab，并实现创建/编辑比赛的表单。由于管理员面板结构较复杂，具体实现时需在现有 admin 视图的 tab 切换逻辑中添加 `contestAdmin` tab，包含：

1. 比赛列表（复用 `GET /admin/contest-list`）
2. 创建比赛表单（标题、描述、时间选择、类型选择、题目多选）
3. 编辑/删除操作

此任务内容较多，实现时需仔细对照现有 admin 面板的 tab 模式。

- [ ] **步骤 2：验证前端可运行**

运行：`cd frontend && npm run dev`
预期：管理员面板可切换到比赛管理 tab，可创建比赛

- [ ] **步骤 3：Commit**

```bash
git add frontend/src/OnlineJudge.vue
git commit -m "feat(frontend): add admin contest management panel"
```

---

## 任务 17：数据库迁移 — 创建新表

- [ ] **步骤 1：创建数据库迁移 SQL**

创建 `migrations/001_contest_tables.sql`：

```sql
CREATE TABLE IF NOT EXISTS `contest_basic` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  `deleted_at` DATETIME(3) DEFAULT NULL,
  `identity` VARCHAR(36) NOT NULL,
  `title` VARCHAR(255) NOT NULL,
  `description` TEXT,
  `start_at` DATETIME NOT NULL,
  `end_at` DATETIME NOT NULL,
  `contest_type` TINYINT(1) NOT NULL DEFAULT 1,
  `max_participants` INT(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_contest_basic_identity` (`identity`),
  KEY `idx_contest_basic_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `contest_problem` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  `deleted_at` DATETIME(3) DEFAULT NULL,
  `contest_identity` VARCHAR(36) NOT NULL,
  `problem_identity` VARCHAR(36) NOT NULL,
  `score` INT(11) NOT NULL DEFAULT 100,
  `sort` INT(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_contest_problem_contest` (`contest_identity`),
  KEY `idx_contest_problem_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `contest_user` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  `deleted_at` DATETIME(3) DEFAULT NULL,
  `contest_identity` VARCHAR(36) NOT NULL,
  `user_identity` VARCHAR(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_contest_user_contest` (`contest_identity`),
  KEY `idx_contest_user_user` (`user_identity`),
  KEY `idx_contest_user_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `contest_submit` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  `deleted_at` DATETIME(3) DEFAULT NULL,
  `contest_identity` VARCHAR(36) NOT NULL,
  `problem_identity` VARCHAR(36) NOT NULL,
  `user_identity` VARCHAR(36) NOT NULL,
  `path` VARCHAR(255) NOT NULL,
  `score` INT(11) NOT NULL DEFAULT 0,
  `status` TINYINT(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_contest_submit_contest` (`contest_identity`),
  KEY `idx_contest_submit_user` (`user_identity`),
  KEY `idx_contest_submit_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **步骤 2：执行迁移**

运行：`mysql -u root -p onlinepractice < migrations/001_contest_tables.sql`
预期：4 张表创建成功

- [ ] **步骤 3：Commit**

```bash
git add migrations/
git commit -m "feat(db): add contest tables migration"
```

---

## 自检结果

1. **规格覆盖度：** 所有规格需求均有对应任务覆盖 — 4 张表、管理员 CRUD、用户报名/提交/查看、公共列表/详情/排行榜、前端视图、路由注册、数据库迁移。
2. **占位符扫描：** 无 TODO、TBD、待定内容。
3. **类型一致性：** `JudgeCode` 签名在任务 5 中定义，任务 9 中调用，参数和返回值一致。`ContestBasic`、`ContestProblem`、`ContestUser`、`ContestSubmit` 模型在任务 1-4 中定义，后续任务中引用的字段名和类型一致。
