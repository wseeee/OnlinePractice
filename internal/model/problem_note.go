package model

import "gorm.io/gorm"

type ProblemNote struct {
	gorm.Model
	Identity        string `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
	UserIdentity    string `gorm:"column:user_identity;type:varchar(36);index" json:"user_identity"`
	ProblemIdentity string `gorm:"column:problem_identity;type:varchar(36);index" json:"problem_identity"`
	Content         string `gorm:"column:content;type:text" json:"content"`
}

func (ProblemNote) TableName() string {
	return "problem_note"
}

func UpsertNote(identity, userIdentity, problemIdentity, content string) error {
	var existing ProblemNote
	err := DB.Where("user_identity = ? AND problem_identity = ?", userIdentity, problemIdentity).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return DB.Create(&ProblemNote{
			Identity:        identity,
			UserIdentity:    userIdentity,
			ProblemIdentity: problemIdentity,
			Content:         content,
		}).Error
	}
	if err != nil {
		return err
	}
	return DB.Model(&existing).Update("content", content).Error
}

func GetNote(userIdentity, problemIdentity string) (*ProblemNote, error) {
	var note ProblemNote
	err := DB.Where("user_identity = ? AND problem_identity = ?", userIdentity, problemIdentity).First(&note).Error
	return &note, err
}

func DeleteNote(userIdentity, problemIdentity string) error {
	return DB.Where("user_identity = ? AND problem_identity = ?", userIdentity, problemIdentity).
		Delete(&ProblemNote{}).Error
}
