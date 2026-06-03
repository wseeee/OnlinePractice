package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/config"
	"OnlinePrictice/internal/service"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ======================== 管理员接口 ========================

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

	startAt, err := time.Parse(config.DateLayout, startAtStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "开始时间格式错误"})
		return
	}
	endAt, err := time.Parse(config.DateLayout, endAtStr)
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

	identity := helper.GetUUID()
	contest := &model.ContestBasic{
		Identity:        identity,
		Title:           title,
		Description:     description,
		StartAt:         startAt,
		EndAt:           endAt,
		ContestType:     contestType,
		MaxParticipants: maxParticipants,
	}

	contestProblems := make([]*model.ContestProblem, 0, len(problemIdentities))
	for i, pi := range problemIdentities {
		contestProblems = append(contestProblems, &model.ContestProblem{
			ContestIdentity: identity,
			ProblemIdentity: pi,
			Score:           100,
			Sort:            i + 1,
		})
	}
	contest.ContestProblems = contestProblems

	err = model.CreateContest(contest).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{"identity": identity},
	})
}

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

	startAt, err := time.Parse(config.DateLayout, startAtStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "开始时间格式错误"})
		return
	}
	endAt, err := time.Parse(config.DateLayout, endAtStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "结束时间格式错误"})
		return
	}

	var problemIdentities []string
	if err := json.Unmarshal([]byte(problemIdentitiesStr), &problemIdentities); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "题目列表格式错误"})
		return
	}

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		contest := &model.ContestBasic{
			Title:           title,
			Description:     description,
			StartAt:         startAt,
			EndAt:           endAt,
			ContestType:     contestType,
			MaxParticipants: maxParticipants,
		}
		err := tx.Model(new(model.ContestBasic)).
			Where("identity = ?", identity).
			Updates(contest).Error
		if err != nil {
			return err
		}
		err = tx.Where("contest_identity = ?", identity).Delete(&model.ContestProblem{}).Error
		if err != nil {
			return err
		}
		contestProblems := make([]*model.ContestProblem, 0, len(problemIdentities))
		for i, pi := range problemIdentities {
			contestProblems = append(contestProblems, &model.ContestProblem{
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
	size, err := strconv.Atoi(c.DefaultQuery("size", config.DefaultSize))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数 size 类型错误"})
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", config.DefaultPage))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数 page 类型错误"})
		return
	}
	page = (page - 1) * size
	keyword := c.Query("keyword")

	list := make([]*model.ContestBasic, 0)
	var count int64
	err = model.GetContestList(keyword).Count(&count).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{"list": list, "count": count},
	})
}

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
	err := model.DeleteContest(identity).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// ======================== 用户接口 ========================

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
	userClaim := u.(*helper.UserJwt)

	var contest model.ContestBasic
	err := model.DB.Where("identity = ?", contestIdentity).First(&contest).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "比赛不存在"})
		return
	}
	if contest.ContestType != 2 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "该比赛无需报名"})
		return
	}

	if model.IsContestUser(contestIdentity, userClaim.Identity) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "已报名，请勿重复操作"})
		return
	}

	if contest.MaxParticipants > 0 {
		count := model.GetContestUserCount(contestIdentity)
		if count >= int64(contest.MaxParticipants) {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "报名人数已满"})
			return
		}
	}

	cu := &model.ContestUser{
		ContestIdentity: contestIdentity,
		UserIdentity:    userClaim.Identity,
	}
	err = model.DB.Create(cu).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "报名失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "报名成功"})
}

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
	language := c.DefaultQuery("language", "go")
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
	userClaim := u.(*helper.UserJwt)

	// 校验比赛状态
	var contest model.ContestBasic
	err = model.DB.Where("identity = ?", contestIdentity).First(&contest).Error
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
	if contest.ContestType == 2 && !model.IsContestUser(contestIdentity, userClaim.Identity) {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "请先报名"})
		return
	}

	// 校验题目是否属于该比赛
	var cp model.ContestProblem
	err = model.DB.Where("contest_identity = ? AND problem_identity = ?", contestIdentity, problemIdentity).First(&cp).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "该题目不属于此比赛"})
		return
	}

	// 保存代码 + 校验
	codePath, valid, err := service.SaveAndValidateCodeLang(code, language)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
		return
	}

	submitIdentity := helper.GetUUID()

	if !valid {
		// 非法代码：同步返回 status=6，不入队
		cs := &model.ContestSubmit{
			Identity:        submitIdentity,
			ContestIdentity: contestIdentity,
			ProblemIdentity: problemIdentity,
			UserIdentity:    userClaim.Identity,
			Path:            codePath,
			Status:          6,
		}
		model.DB.Create(cs)
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": map[string]interface{}{
				"submit_identity": submitIdentity,
				"status":          6,
				"msg":             "无效代码",
			},
		})
		return
	}

	// 创建记录 status=0，不增加全站 submit_num
	cs := &model.ContestSubmit{
		Identity:        submitIdentity,
		ContestIdentity: contestIdentity,
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity,
		Path:            codePath,
		Status:          0,
	}
	if err := model.DB.Create(cs).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "提交保存失败"})
		return
	}

	// 发布到 MQ
	task := &model.JudgeTask{
		SubmitIdentity:  submitIdentity,
		RecordType:      "contest",
		CodePath:        codePath,
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity,
		IsContest:       true,
		ContestIdentity: contestIdentity,
		ProblemScore:    cp.Score,
	}
	if err := model.PublishJudgeTask(model.RoutingKeyContest, task); err != nil {
		// RabbitMQ 不可用时降级为同步判题
		passed, total, status, judgeErr := service.JudgeCodeWithLang(codePath, problemIdentity, language)
		if judgeErr != nil {
			model.DB.Model(cs).Update("status", 6)
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "判题失败: " + judgeErr.Error()})
			return
		}
		score := 0
		if total > 0 {
			score = passed * cp.Score / total
		}
		model.DB.Model(cs).Updates(map[string]interface{}{
			"status": status,
			"score":  score,
		})
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": map[string]interface{}{
				"submit_identity": submitIdentity,
				"status":          status,
				"score":           score,
				"msg":             service.StatusMsg(status),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"submit_identity": submitIdentity,
			"status":          0,
			"msg":             "已提交，排队中",
		},
	})
}

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
	userClaim := u.(*helper.UserJwt)

	size, err := strconv.Atoi(c.DefaultQuery("size", config.DefaultSize))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数 size 类型错误"})
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", config.DefaultPage))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数 page 类型错误"})
		return
	}
	page = (page - 1) * size

	list := make([]*model.ContestSubmit, 0)
	var count int64
	err = model.GetContestSubmitList(contestIdentity, problemIdentity, userClaim.Identity).
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

// ======================== 公共接口 ========================

// GetContestListPublic
// @Tags 公共方法
// @Summary 比赛列表
// @Param page query int false "page"
// @Param size query int false "size"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /contest-list [get]
func GetContestListPublic(c *gin.Context) {
	size, err := strconv.Atoi(c.DefaultQuery("size", config.DefaultSize))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数 size 类型错误"})
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", config.DefaultPage))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数 page 类型错误"})
		return
	}
	page = (page - 1) * size

	list := make([]*model.ContestBasic, 0)
	var count int64
	err = model.DB.Model(new(model.ContestBasic)).
		Order("start_at DESC").
		Count(&count).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}

	type contestInfo struct {
		*model.ContestBasic
		ParticipantCount int64 `json:"participant_count"`
		ContestStatus    int   `json:"contest_status"`
	}
	result := make([]contestInfo, len(list))
	now := time.Now()
	for i, item := range list {
		status := 0
		if now.After(item.StartAt) && now.Before(item.EndAt) {
			status = 1
		} else if now.After(item.EndAt) {
			status = 2
		}
		result[i] = contestInfo{
			ContestBasic:     item,
			ParticipantCount: model.GetContestUserCount(item.Identity),
			ContestStatus:    status,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{"list": result, "count": count},
	})
}

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
	var contest model.ContestBasic
	err := model.GetContestDetail(identity).First(&contest).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "比赛不存在"})
		return
	}

	participantCount := model.GetContestUserCount(identity)
	now := time.Now()
	status := 0
	if now.After(contest.StartAt) && now.Before(contest.EndAt) {
		status = 1
	} else if now.After(contest.EndAt) {
		status = 2
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

	size, err := strconv.Atoi(c.DefaultQuery("size", config.DefaultSize))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数 size 类型错误"})
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", config.DefaultPage))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "参数 page 类型错误"})
		return
	}

	type userRank struct {
		UserIdentity string    `json:"user_identity"`
		UserName     string    `json:"user_name"`
		TotalScore   int       `json:"total_score"`
		TotalTime    time.Time `json:"total_time"`
	}

	// 尝试从 Redis 缓存获取
	cacheKey := "contest_rank:" + contestIdentity
	cached, err := model.RDB.Get(config.CTX, cacheKey).Result()
	if err == nil && cached != "" {
		var ranks []userRank
		if json.Unmarshal([]byte(cached), &ranks) == nil {
			total := len(ranks)
			start := (page - 1) * size
			end := start + size
			if start > total {
				start = total
			}
			if end > total {
				end = total
			}
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"data": map[string]interface{}{"list": ranks[start:end], "count": total},
			})
			return
		}
	}

	// 查询该比赛所有提交
	var submits []model.ContestSubmit
	err = model.DB.Where("contest_identity = ?", contestIdentity).
		Find(&submits).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}

	type userProblemScore struct {
		Score       int
		FirstACTime time.Time
	}

	userMap := make(map[string]map[string]*userProblemScore)
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

	// 批量查询用户名
	userIDs := make([]string, 0, len(userMap))
	for uid := range userMap {
		userIDs = append(userIDs, uid)
	}
	var users []model.UserBasic
	model.DB.Where("identity IN ?", userIDs).Find(&users)
	nameMap := make(map[string]string, len(users))
	for _, u := range users {
		nameMap[u.Identity] = u.Name
	}

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
		ranks = append(ranks, userRank{
			UserIdentity: userIdentity,
			UserName:     nameMap[userIdentity],
			TotalScore:   totalScore,
			TotalTime:    earliestAC,
		})
	}

	sort.Slice(ranks, func(i, j int) bool {
		if ranks[i].TotalScore != ranks[j].TotalScore {
			return ranks[i].TotalScore > ranks[j].TotalScore
		}
		return ranks[i].TotalTime.Before(ranks[j].TotalTime)
	})

	// 缓存到 Redis（30 秒）
	rankJSON, _ := json.Marshal(ranks)
	model.RDB.Set(config.CTX, cacheKey, string(rankJSON), 30*time.Second)

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

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  pagedRanks,
			"count": total,
		},
	})
}
