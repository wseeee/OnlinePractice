package model

import (
	"gorm.io/gorm"
	uuid "github.com/satori/go.uuid"
)

type Tag struct {
	gorm.Model
	Identity string `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
	Name     string `gorm:"column:name;type:varchar(64);uniqueIndex" json:"name"`
	Color    string `gorm:"column:color;type:varchar(7);default:#409eff" json:"color"` // hex color
}

func (Tag) TableName() string {
	return "tag"
}

type ProblemTag struct {
	gorm.Model
	Identity        string `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
	ProblemIdentity string `gorm:"column:problem_identity;type:varchar(36);index" json:"problem_identity"`
	TagIdentity     string `gorm:"column:tag_identity;type:varchar(36);index" json:"tag_identity"`
	Tag             *Tag   `gorm:"foreignKey:tag_identity;references:identity" json:"tag"`
}

func (ProblemTag) TableName() string {
	return "problem_tag"
}

// Tag CRUD
func ListTags() ([]*Tag, error) {
	var tags []*Tag
	err := DB.Order("name ASC").Find(&tags).Error
	return tags, err
}

func CreateTag(t *Tag) error {
	return DB.Create(t).Error
}

func UpdateTag(identity, name, color string) error {
	return DB.Model(new(Tag)).Where("identity = ?", identity).
		Updates(map[string]interface{}{"name": name, "color": color}).Error
}

func DeleteTag(identity string) error {
	// 同时删除关联
	DB.Where("tag_identity = ?", identity).Delete(&ProblemTag{})
	return DB.Where("identity = ?", identity).Delete(&Tag{}).Error
}

// Problem-Tag 关联
func SetProblemTags(problemIdentity string, tagIdentities []string) error {
	DB.Where("problem_identity = ?", problemIdentity).Delete(&ProblemTag{})
	for _, ti := range tagIdentities {
		DB.Create(&ProblemTag{
			Identity:        uuid.NewV4().String(),
			ProblemIdentity: problemIdentity,
			TagIdentity:     ti,
		})
	}
	return nil
}

func GetProblemTags(problemIdentity string) ([]*ProblemTag, error) {
	var tags []*ProblemTag
	err := DB.Where("problem_identity = ?", problemIdentity).
		Preload("Tag").Find(&tags).Error
	return tags, err
}
