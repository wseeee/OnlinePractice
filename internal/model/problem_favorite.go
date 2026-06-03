package model

import "gorm.io/gorm"

type ProblemFavorite struct {
	gorm.Model
	Identity        string        `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
	UserIdentity    string        `gorm:"column:user_identity;type:varchar(36);index" json:"user_identity"`
	ProblemIdentity string        `gorm:"column:problem_identity;type:varchar(36);index" json:"problem_identity"`
	ProblemBasic    *ProblemBasic `gorm:"foreignKey:problem_identity;references:identity" json:"problem_basic"`
}

func (ProblemFavorite) TableName() string {
	return "problem_favorite"
}

func AddFavorite(identity, userIdentity, problemIdentity string) error {
	return DB.Create(&ProblemFavorite{
		Identity:        identity,
		UserIdentity:    userIdentity,
		ProblemIdentity: problemIdentity,
	}).Error
}

func RemoveFavorite(userIdentity, problemIdentity string) error {
	return DB.Where("user_identity = ? AND problem_identity = ?", userIdentity, problemIdentity).
		Delete(&ProblemFavorite{}).Error
}

func IsFavorited(userIdentity, problemIdentity string) bool {
	var count int64
	DB.Model(new(ProblemFavorite)).
		Where("user_identity = ? AND problem_identity = ?", userIdentity, problemIdentity).
		Count(&count)
	return count > 0
}

func GetFavoriteProblems(userIdentity string, page, size int) ([]*ProblemFavorite, int64, error) {
	var list []*ProblemFavorite
	var count int64
	q := DB.Model(new(ProblemFavorite)).Where("user_identity = ?", userIdentity)
	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Offset((page-1)*size).Limit(size).
		Preload("ProblemBasic", func(db *gorm.DB) *gorm.DB { return db.Omit("content") }).
		Find(&list).Error
	return list, count, err
}

