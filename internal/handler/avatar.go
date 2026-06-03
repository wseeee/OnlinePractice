package handler

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var avatarExts = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

const avatarMaxSize = 2 * 1024 * 1024 // 2MB

// PostAvatarUpload 上传头像（用户 JWT）
func PostAvatarUpload(c *gin.Context) {
	if model.Store == nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "对象存储未配置"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "请选择文件"})
		return
	}

	if file.Size > avatarMaxSize {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "头像最大 2MB"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	ct, ok := avatarExts[ext]
	if !ok {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "仅支持 jpg/jpeg/png/webp"})
		return
	}

	u, _ := c.Get("user_claims")
	userIdentity := u.(*helper.UserJwt).Identity
	key := fmt.Sprintf("users/%s/avatar%s", userIdentity, ext)

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "文件读取失败"})
		return
	}
	defer f.Close()

	buf := make([]byte, file.Size)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "文件读取失败"})
		return
	}
	buf = buf[:n]

	bucket := os.Getenv("MINIO_BUCKET_AVATAR")
	if bucket == "" {
		bucket = "oj-avatar"
	}

	if err := model.Store.Put(c.Request.Context(), bucket, key, buf, ct); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "上传失败"})
		return
	}

	// 更新用户 avatar_key
	if err := model.DB.Model(&model.UserBasic{}).Where("identity = ?", userIdentity).Update("avatar_key", key).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新头像记录失败"})
		return
	}

	// 生成预签名 URL
	url, err := model.Store.PresignedGetURL(c.Request.Context(), bucket, key, time.Hour)
	if err != nil {
		// 上传成功但预签名失败，仍返回成功但 URL 为空
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"avatar_key": key,
				"avatar_url": "",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"avatar_key": key,
			"avatar_url": url,
		},
	})
}

// GetMyAvatarURL 获取当前用户头像 URL（用户 JWT）
func GetMyAvatarURL(c *gin.Context) {
	u, _ := c.Get("user_claims")
	userIdentity := u.(*helper.UserJwt).Identity
	getAvatarURL(c, userIdentity)
}

// GetAvatarURL 获取任意用户头像 URL（公共）
func GetAvatarURL(c *gin.Context) {
	userIdentity := c.Query("user_identity")
	if userIdentity == "" {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "缺少 user_identity"})
		return
	}
	getAvatarURL(c, userIdentity)
}

func getAvatarURL(c *gin.Context, userIdentity string) {
	if model.Store == nil {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": nil})
		return
	}

	var user model.UserBasic
	if err := model.DB.Where("identity = ?", userIdentity).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户不存在"})
		return
	}

	if user.AvatarKey == "" {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": nil})
		return
	}

	bucket := os.Getenv("MINIO_BUCKET_AVATAR")
	if bucket == "" {
		bucket = "oj-avatar"
	}

	url, err := model.Store.PresignedGetURL(c.Request.Context(), bucket, user.AvatarKey, time.Hour)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "获取头像URL失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"avatar_url": url,
		},
	})
}
