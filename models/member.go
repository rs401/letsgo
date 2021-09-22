package models

import "gorm.io/gorm"

type Member struct {
	gorm.Model
	ForumID uint `json:"ForumID" gorm:"not null"`
	Forum   Forum
	UserID  uint `json:"UserID" gorm:"not null"`
	User    User
}

type PendingMember struct {
	gorm.Model
	ForumID uint `json:"ForumID" gorm:"not null"`
	Forum   Forum
	UserID  uint `json:"UserID" gorm:"not null"`
	User    User
}

// func IsMember(fid, uid uint) bool {
// 	var members []Member
// 	DBConn.Where("forum_id = ?", fid).Find(&members)
// 	for _, member := range members {
// 		if member.UserID == uid {
// 			return true
// 		}
// 	}
// 	return false
// }
