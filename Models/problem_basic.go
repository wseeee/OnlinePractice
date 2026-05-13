package Models

import (
	"gorm.io/gorm"
)

type ProblemBasic struct {
	gorm.Model
	Identity          string             `gorm:"column:identity;type:varchar(36);" json:"identity"`
	ProblemCategories []*ProblemCategory `gorm:"foreignKey:problem_id;references:id" json:"problem_categories"`
	Title             string             `gorm:"column:title;type:varchar(255);" json:"title"`
	Content           string             `gorm:"column:content;type:text;" json:"content"`
	MaxMem            int                `gorm:"column:max_mem;type:int(11);" json:"max_mem"`
	MaxRuntime        int                `gorm:"column:max_runtime;type:int(11);" json:"max_runtime"`
	SubmitNum         int                `gorm:"column:submit_num;type:int(11);" json:"submit_num"`
	PassNum           int                `gorm:"column:pass_num;type:int(11);" json:"pass_num"`
	TestCases         []*TestCase        `gorm:"foreignKey:problem_identity;references:identity" json:"test_cases"`
}

func (table *ProblemBasic) TableName() string {
	return "problem_basic"
}

func GetProblemList(keyword string, categoryIdentity string) *gorm.DB {

	tx := DB.Model(new(ProblemBasic)).
		Preload("ProblemCategories").
		Preload("ProblemCategories.CategoryBasic").
		Where("title like ? OR content like ?",
			"%"+keyword+"%", "%"+keyword+"%")
	if categoryIdentity != "" {
		tx = tx.Joins("RIGHT JOIN problem_category pc on pc.problem_id=problem_basic.id").
			Where("pc.category_id = ("+
				"SELECT cb.id "+
				"FROM category_basic cb "+
				"WHERE cb.identity = ?)", categoryIdentity)
	}
	return tx
}

func GetProblemDetail(identity string) *gorm.DB {
	return DB.Model(new(ProblemBasic)).
		Preload("ProblemCategories").
		Preload("ProblemCategories.CategoryBasic").
		Where("identity = ?", identity)
}

func CreateProblem(p *ProblemBasic) *gorm.DB {
	return DB.Model(new(ProblemBasic)).Create(p)
}
