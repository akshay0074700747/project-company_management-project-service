package entities

import "time"

type Credentials struct {
	ProjectID       string `gorm:"primaryKey"`
	ProjectUsername string `gorm:"unique"`
	Name            string
	Aim             string
	Description     string
	IsCompanybased  bool
	IsPublic        bool
	Deadline        time.Time
}

type Members struct {
	MemberID     string
	ProjectID    string `gorm:"foreignKey:ProjectID;references:credentials(project_id)"`
	RoleID       uint
	PermissionID uint
	StatusID     uint `gorm:"foreignKey:StatusID;references:member_statuses(id)"`
}

type MemberStatus struct {
	ID     uint   `gorm:"primaryKey"`
	Status string `gorm:"unique"`
}

type Companies struct {
	CompanyID string
	ProjectID string `gorm:"foreignKey:ProjectID;references:credentials(project_id)"`
}

type Owners struct {
	ProjectID string `gorm:"foreignKey:ProjectID;references:credentials(project_id)"`
	OwnerID   string
}

type TaskAssignations struct {
	UserID      string
	ProjectID   string `gorm:"foreignKey:ProjectID;references:credentials(project_id)"`
	Task        string
	Description string
	ObjectName  string `gorm:"unique"`
	Stages      int
	Deadline    time.Time
	IsVerified bool `gorm:"default:false"`
}

type TaskStatuses struct {
	ID     uint   `gorm:"primaryKey"`
	Status string `gorm:"unique"`
	Stat   int
}

type NonTechnicalTaskDetials struct {
	UserID      string
	ProjectID   string `gorm:"foreignKey:ProjectID;references:credentials(project_id)"`
	Task        string
	Description string
}

type Issues struct {
	ID          uint   `gorm:"primaryKey"`
	ProjectID   string `gorm:"foreignKey:ProjectID;references:credentials(project_id)"`
	UserID      string
	Description string
	RaiserID    string
}

type Ratings struct {
	ID        uint   `gorm:"primaryKey"`
	ProjectID string `gorm:"foreignKey:ProjectID;references:credentials(project_id)"`
	UserID    string
	Rating    float32
	Feedback  string
}

type Extensions struct {
	ID          uint   `gorm:"primaryKey"`
	ProjectID   string `gorm:"foreignKey:ProjectID;references:credentials(project_id)"`
	UserID      string
	ExtendTo time.Time
	Description string
	IsAccepted  bool `gorm:"default:false"`
}
