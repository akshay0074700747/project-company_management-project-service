package adapters

import (
	"github.com/akshay0074700747/project-company_management-project-service/entities"
	"gorm.io/gorm"
)

type ProjectAdapter struct {
	DB *gorm.DB
}

func NewProjectAdapter(db *gorm.DB) *ProjectAdapter {
	return &ProjectAdapter{
		DB: db,
	}
}

func (project *ProjectAdapter) CreateProject(req entities.Credentials) (entities.Credentials, error) {

	query := "INSERT INTO credentials (project_username,name,aim,description,is_companybased) VALUES($1,$2,$3,$4,$5) RETURNING project_id,project_username,name,aim,description,is_companybased"
	var res entities.Credentials

	tx := project.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := project.DB.Raw(query, req.ProjectUsername, req.Name, req.Aim, req.Description, req.IsCompanybased).Scan(&res).Error; err != nil {
		tx.Rollback()
		return res, err
	}

	if err := tx.Commit().Error; err != nil {
		return res, err
	}
	return res, nil
}
