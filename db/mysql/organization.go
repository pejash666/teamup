package mysql

import (
	"gorm.io/gorm"
)

type Organization struct {
	gorm.Model
	SportType     string `gorm:"sport_type" json:"sport_type"`                  // 运动类型
	Name          string `gorm:"column:name" json:"name"`                       // 组织名字
	City          string `gorm:"column:city" json:"city"`                       // 城市
	Address       string `gorm:"column:address" json:"address"`                 // 详细地址
	Longitude     string `gorm:"column:longitude" json:"longitude"`             // 经度
	Latitude      string `gorm:"column:latitude" json:"latitude"`               // 纬度
	HostOpenID    string `gorm:"host_open_id" json:"host_open_id"`              // 主理人openID
	Contact       string `gorm:"column:contact" json:"contact"`                 // 联系方式
	Logo          string `gorm:"column:logo" json:"logo"`                       // 组织图标logo
	TotalEventNum int    `gorm:"column:total_event_num" json:"total_event_num"` // 活动次数
	IsApproved    int    `gorm:"column:is_approved" json:"is_approved"`         // 是否通过审批
	Reviewer      string `gorm:"column:reviewer" json:"reviewer"`               // 审批人
	IsTest        int    `gorm:"column:is_test" json:"is_test"`                 // 是否是测试组织
}

func (o Organization) TableName() string {
	return "organization_info"
}

type OrganizationWithoutGorm struct {
	ID            uint   `gorm:"column:id" json:"id"`                           // 组织ID
	SportType     string `gorm:"column:sport_type" json:"sport_type"`           // 运动类型
	Name          string `gorm:"column:name" json:"name"`                       // 组织名字
	City          string `gorm:"column:city" json:"city"`                       // 城市
	Address       string `gorm:"column:address" json:"address"`                 // 详细地址
	HostOpenID    string `gorm:"host_open_id" json:"host_open_id"`              // 主理人openID
	Contact       string `gorm:"column:contact" json:"contact"`                 // 联系方式
	Logo          string `gorm:"column:logo" json:"logo"`                       // 组织图标logo
	TotalEventNum int    `gorm:"column:total_event_num" json:"total_event_num"` // 活动次数
	IsApproved    int    `gorm:"column:is_approved" json:"is_approved"`         // 是否通过审批
	Reviewer      string `gorm:"column:reviewer" json:"reviewer"`               // 审批人
}
