package model

import "gorm.io/gorm"

type CategoryBasic struct {
	gorm.Model
	Identity string `gorm:"column:identity;type:varchar(36);" json:"identity"`
	Name     string `gorm:"column:name;type:varchar(100);" json:"name"`
	ParentId int    `gorm:"column:parent_id;type:int(11);" json:"parent_id"`
}

func (table *CategoryBasic) TableName() string {
	return "category_basic"
}

func GetCategoryList(keyword string) *gorm.DB {
	return DB.Model(new(CategoryBasic)).
		Where("name like ?", "%"+keyword+"%")
}

func CreateCategory(c *CategoryBasic) *gorm.DB {
	return DB.Model(new(CategoryBasic)).Create(c)
}
func DeleteCategory(identity string) *gorm.DB {
	return DB.Model(new(CategoryBasic)).
		Where("identity = ?", identity).Delete(new(CategoryBasic))
}
func UpdateCategory(identity string, c *CategoryBasic) *gorm.DB {
	return DB.Model(new(CategoryBasic)).
		Where("identity = ?", identity).Updates(c)
}
