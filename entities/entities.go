package entities

type Credentials struct {
	ProjectID       string `gorm:"primaryKey"`
	ProjectUsername string `gorm:"unique"`
	Name            string
	Aim             string
	Description     string
	IsCompanybased  bool
}

type Members struct {
	memberID  string
	ProjectID string `gorm:"foreignKey:ProjectID;references:credentials(project_id)"`
}

type Companies struct {
	CompanyID string
	ProjectID string `gorm:"foreignKey:ProjectID;references:credentials(project_id)"`
}
