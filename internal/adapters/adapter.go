package adapters

import (
	"errors"

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

func (project *ProjectAdapter) CreateProject(req entities.Credentials, compID string) (entities.Credentials, error) {

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

	if err := project.insertIntoCompanyBased(entities.Companies{
		CompanyID: compID,
		ProjectID: req.ProjectID,
	}); err != nil {
		tx.Rollback()
		return res, err
	}

	if err := tx.Commit().Error; err != nil {
		return res, err
	}
	return res, nil
}

func (project *ProjectAdapter) insertIntoCompanyBased(req entities.Companies) error {

	query := "INSERT INTO companies (company_id,project_id) VALUES($1,$2)"

	if err := project.DB.Exec(query, req.CompanyID, req.ProjectID).Error; err != nil {
		return err
	}

	return nil
}

func (project *ProjectAdapter) IsProjectUsernameExists(req string) (bool, error) {

	query := "SELECT * FROM credentials WHERE project_username = $1"

	res := project.DB.Exec(query, req)
	if res.Error != nil {
		return true, res.Error
	}

	if res.RowsAffected != 0 {
		return true, nil
	}

	return false, nil
}

func (project *ProjectAdapter) AddMember(req entities.Members) error {

	query := "INSERT INTO members (member_id,project_id,role_id,permission_id) VALUES($1,$2,$3,$4)"

	tx := project.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := project.DB.Raw(query, req.MemberID, req.ProjectID, req.RoleID, req.PermissionID).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (project *ProjectAdapter) IsMemberExists(id string) (bool, error) {

	query := "SELECT * FROM members WHERE member_id = $1"

	res := project.DB.Exec(query, id)
	if res.Error != nil {
		return true, res.Error
	}

	if res.RowsAffected != 0 {
		return true, nil
	}

	return false, nil
}

func (project *ProjectAdapter) AcceptInvitation(req entities.Members) error {

	query := "UPDATE members SET is_accepted = true WHERE member_id = $1 AND project_id = $2"

	res := project.DB.Exec(query, req.MemberID, req.ProjectID)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return errors.New("the project id or userId doesnt exist")
	}

	return nil
}

func (comp *ProjectAdapter) GetProjectInvites(memID string) ([]entities.ProjectInviteUsecase, error) {

	query := "SELECT m.project_id, COUNT(*) AS accepted_members FROM members m JOIN members ma ON m.project_id = ma.project_id AND ma.member_id = $1 WHERE ma.is_accepted = TRUE GROUP BY m.project_id"
	var res []entities.ProjectInviteUsecase

	if err := comp.DB.Raw(query, memID).Scan(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}
