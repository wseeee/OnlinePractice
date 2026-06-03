package model

import "gorm.io/gorm"

type Notification struct {
	gorm.Model
	Identity        string `gorm:"column:identity;type:varchar(36);uniqueIndex" json:"identity"`
	UserIdentity    string `gorm:"column:user_identity;type:varchar(36);index" json:"user_identity"`
	Type            string `gorm:"column:type;type:varchar(32)" json:"type"` // comment_reply, contest_reminder, system_announcement
	Title           string `gorm:"column:title;type:varchar(255)" json:"title"`
	Content         string `gorm:"column:content;type:text" json:"content"`
	RelatedIdentity string `gorm:"column:related_identity;type:varchar(36)" json:"related_identity"` // 关联资源 identity
	IsRead          bool   `gorm:"column:is_read;type:tinyint(1);default:0" json:"is_read"`
}

func (Notification) TableName() string {
	return "notification"
}

func CreateNotification(n *Notification) error {
	return DB.Create(n).Error
}

func GetNotificationsByUser(userIdentity string, page, size int) ([]*Notification, int64, error) {
	var list []*Notification
	var count int64
	q := DB.Model(new(Notification)).Where("user_identity = ?", userIdentity)
	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, count, err
}

func MarkNotificationRead(identity, userIdentity string) error {
	return DB.Model(new(Notification)).
		Where("identity = ? AND user_identity = ?", identity, userIdentity).
		Update("is_read", true).Error
}

func MarkAllNotificationsRead(userIdentity string) error {
	return DB.Model(new(Notification)).
		Where("user_identity = ? AND is_read = 0", userIdentity).
		Update("is_read", true).Error
}

func GetUnreadNotificationCount(userIdentity string) (int64, error) {
	var count int64
	err := DB.Model(new(Notification)).
		Where("user_identity = ? AND is_read = 0", userIdentity).
		Count(&count).Error
	return count, err
}
