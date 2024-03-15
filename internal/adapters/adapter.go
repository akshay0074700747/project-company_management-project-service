package adapters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/akshay0074700747/project-company_management-project-service/entities"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type ProjectAdapter struct {
	DB      *gorm.DB
	MinioDB *minio.Client
}

func NewProjectAdapter(db *gorm.DB, minioDB *minio.Client) *ProjectAdapter {
	return &ProjectAdapter{
		DB:      db,
		MinioDB: minioDB,
	}
}

func (project *ProjectAdapter) CreateProject(req entities.Credentials, compID string) (entities.Credentials, error) {

	query := "INSERT INTO credentials (project_id,project_username,name,aim,description,is_companybased,is_public,deadline) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING project_id,project_username,name,aim,description,is_companybased,is_public"
	var res entities.Credentials

	tx := project.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := project.DB.Raw(query, req.ProjectID, req.ProjectUsername, req.Name, req.Aim, req.Description, req.IsCompanybased, req.IsPublic, req.Deadline).Scan(&res).Error; err != nil {
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

	query := "INSERT INTO members (member_id,project_id,role_id,permission_id,status_id) VALUES($1,$2,$3,$4,(SELECT id FROM member_statuses WHERE status = 'PENDING'))"

	tx := project.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := project.DB.Exec(query, req.MemberID, req.ProjectID, req.RoleID, req.PermissionID).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (project *ProjectAdapter) IsMemberExists(id string, projID string) (bool, error) {

	query := "SELECT * FROM members INNER JOIN member_statuses ON status_id = id AND status = 'ACCEPTED' WHERE member_id = $1 AND project_id = $2"

	res := project.DB.Exec(query, id, projID)
	if res.Error != nil {
		return true, res.Error
	}

	if res.RowsAffected != 0 {
		return true, nil
	}

	return false, nil
}

func (project *ProjectAdapter) AcceptInvitation(req entities.Members) error {

	query := "UPDATE members SET status_id = (SELECT id FROM member_statuses WHERE status = 'ACCEPTED') WHERE member_id = $1 AND project_id = $2 "

	res := project.DB.Exec(query, req.MemberID, req.ProjectID)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return errors.New("the project id or userId doesnt exist")
	}

	return nil
}

func (project *ProjectAdapter) DenyInvitation(req entities.Members) error {

	query := "UPDATE members SET status_id = (SELECT id FROM member_statuses WHERE status = 'REJECTED') WHERE member_id = $1 AND project_id = $2 "

	res := project.DB.Exec(query, req.MemberID, req.ProjectID)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return errors.New("the project id or userId doesnt exist")
	}

	return nil
}

func (project *ProjectAdapter) GetProjectInvites(memID string) ([]entities.ProjectInviteUsecase, error) {

	query := "SELECT m.project_id, COUNT(*) AS accepted_members FROM members m JOIN members ma ON m.project_id = ma.project_id AND ma.member_id = $1 GROUP BY m.project_id"
	var res []entities.ProjectInviteUsecase

	if err := project.DB.Raw(query, memID).Scan(&res).Error; err != nil {
		return nil, err
	}
	var ex string
	for i, v := range res {
		if err := project.DB.Raw("SELECT status FROM member_statuses INNER JOIN members ON member_statuses.id = members.status_id WHERE members.member_id = $1 AND members.project_id = $2", memID, v.ProjectID).Scan(&ex).Error; err != nil {
			return res, err
		}
		if ex == "ACCEPTED" {
			res = append(res[:i], res[i+1:]...)
		}
	}
	return res, nil
}

func (project *ProjectAdapter) AddOwner(projectId, ownerID string) error {

	query := "INSERT INTO owners (project_id,owner_id) VALUES($1,$2)"

	if err := project.DB.Exec(query, projectId, ownerID).Error; err != nil {
		return err
	}

	return nil
}

func (project *ProjectAdapter) GetNoofMembers(projectID string) (uint, error) {

	query := "SELECT COUNT(*) FROM members WHERE project_id = $1 GROUP BY project_id"
	var res uint

	if err := project.DB.Raw(query, projectID).Scan(&res).Error; err != nil {
		return res, err
	}

	return res, nil
}

func (project *ProjectAdapter) GetProjectDetails(projectID string) (entities.Credentials, error) {

	query := "SELECT * FROM credentials WHERE project_id = $1"
	var res entities.Credentials

	if err := project.DB.Raw(query, projectID).Scan(&res).Error; err != nil {
		return entities.Credentials{}, err
	}

	return res, nil
}

func (project *ProjectAdapter) GetProjectMembers(projID string) ([]entities.GetProjectMembersUsecase, error) {

	query := "SELECT member_id AS user_id,role_id,permission_id FROM members WHERE project_id = $1"
	var res []entities.GetProjectMembersUsecase

	if err := project.DB.Raw(query, projID).Scan(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func (project *ProjectAdapter) AddMemberStatueses(status string) error {

	query := "INSERT INTO member_statuses (status) VALUES($1)"
	if err := project.DB.Exec(query, status).Error; err != nil {
		return err
	}

	return nil
}

func (project *ProjectAdapter) InsertTasktoMinio(ctx context.Context, fileName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) error {

	_, err := project.MinioDB.PutObject(ctx, "tasks-storage-bucket", fileName, reader, objectSize, opts)
	if err != nil {
		return err
	}

	return nil
}

func (project *ProjectAdapter) InsertTaskDetails(req entities.TaskAssignations) error {

	query := "INSERT INTO task_assignations (user_id,project_id,task,description,object_name,stages,deadline) VALUES($1,$2,$3,$4,$5,$6,$7)"
	if err := project.DB.Exec(query, req.UserID, req.ProjectID, req.Task, req.Description, req.ObjectName, req.Stages, req.Deadline).Error; err != nil {
		return err
	}

	return nil
}

func (project *ProjectAdapter) GetTaskFromMinio(ctx context.Context, objectName string, opts minio.GetObjectOptions) ([]byte, error) {

	object, err := project.MinioDB.GetObject(ctx, "tasks-storage-bucket", objectName, opts)
	if err != nil {
		return nil, err
	}
	info, _ := object.Stat()
	size := info.Size
	var res = make([]byte, size)
	n, err := object.Read(res)
	if err != nil && err != io.EOF {
		return nil, err
	}

	fmt.Println(n)
	fmt.Println(string(res))

	return res, nil
}

func (project *ProjectAdapter) GetTaskDetails(userID, projectID string) (entities.TaskAssignations, error) {

	var res entities.TaskAssignations
	query := "SELECT * FROM task_assignations WHERE user_id = $1 AND project_id = $2 LIMIT 1"

	fmt.Println(res)
	if err := project.DB.Raw(query, userID, projectID).Scan(&res).Error; err != nil {
		return entities.TaskAssignations{}, err
	}
	return res, nil
}

func (project *ProjectAdapter) InsertStatuses(req entities.TaskStatuses) error {

	query := "INSERT INTO task_statuses (status,stat) VALUES($1,$2)"
	if err := project.DB.Exec(query, req.Status, req.Stat).Error; err != nil {
		return err
	}

	return nil
}

func (project *ProjectAdapter) GetRoleID(usrID string) (uint, error) {

	query := "SELECT role_id from members WHERE member_id = $1"
	var role uint

	if err := project.DB.Raw(query, usrID).Scan(&role).Error; err != nil {
		return 0, err
	}

	return role, nil
}

func (project *ProjectAdapter) LogintoProject(projectID, memberID string) (entities.Members, error) {

	query := "SELECT project_id,role_id,permission_id from members WHERE project_id = $1 AND member_id = $2"
	var res entities.Members

	if err := project.DB.Raw(query, projectID, memberID).Scan(&res).Error; err != nil {
		return entities.Members{}, err
	}

	return res, nil
}

func (project *ProjectAdapter) GetIDfromName(usrName string) (string, error) {

	query := "SELECT project_id FROM credentials WHERE project_username = $1"
	var res string

	tx := project.DB.Raw(query, usrName).Scan(&res)
	if tx.Error != nil {
		return "", tx.Error
	}

	if tx.RowsAffected == 0 {
		return "", errors.New("the username doesnt Exist")
	}

	return res, nil
}

func (project *ProjectAdapter) GetStagesandDeadline(users []string, projectId string) ([]entities.TaskAssignations, error) {

	var res []entities.TaskAssignations

	if err := project.DB.Model(&entities.TaskAssignations{}).Where("user_id IN ? AND project_id = ?", users, projectId).Scan(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func (project *ProjectAdapter) GetTaskStatuses(progresses []int) ([]string, error) {

	var statuses []string
	if err := project.DB.Model(&entities.TaskStatuses{}).Where("stat IN ?", progresses).Pluck("status", &statuses).Error; err != nil {
		return nil, err
	}

	return statuses, nil
}

func (project *ProjectAdapter) GetListofRoleIds(users []string, projectID string) ([]uint, error) {

	var roleIds []uint
	if err := project.DB.Model(&entities.Members{}).Where("member_id IN ? AND project_id = ?", users, projectID).Pluck("role_id", &roleIds).Error; err != nil {
		return nil, err
	}

	return roleIds, nil
}

func (project *ProjectAdapter) InsertNonTechnicalTasks(req entities.NonTechnicalTaskDetials) error {

	query := "INSERT INTO non_technical_task_detials (project_id,user_id,task,description) VALUES($1,$2,$3,$4)"
	if err := project.DB.Exec(query, req.ProjectID, req.UserID, req.Task, req.Description).Error; err != nil {
		return err
	}

	return nil
}

func (project *ProjectAdapter) GetCountofLivemembers(projectID string) (int, error) {

	query := "SELECT count(*) AS count FROM members WHERE project_id = $1 GROUP BY project_id "
	var count int

	if err := project.DB.Raw(query, projectID).Scan(&count).Error; err != nil {
		return 0, nil
	}

	return count, nil
}

func (project *ProjectAdapter) GetProjectDeadline(projectID string) (time.Time, error) {

	query := "SELECT deadline FROM credentials WHERE project_id = $1"
	var res time.Time

	if err := project.DB.Raw(query, projectID).Scan(&res).Error; err != nil {
		return time.Time{}, err
	}

	return res, nil
}

func (project *ProjectAdapter) GetStagesofProgress(projectID, userID string) (int, error) {

	query := "SELECT stages FROM task_assignations WHERE project_id = $1 AND user_id = $2"
	var res int

	if err := project.DB.Raw(query, projectID, userID).Scan(&res).Error; err != nil {
		return 0, err
	}

	return res, nil
}

func (project *ProjectAdapter) IsOwner(user_id, project_id string) (bool, error) {

	query := "SELECT * FROM owners WHERE project_id = $1 AND owner_id = $2"
	tx := project.DB.Exec(query, project_id, user_id)
	if tx.Error != nil {
		return false, tx.Error
	}

	if tx.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

func (proj *ProjectAdapter) IsCompanyBased(projectID string) (bool, string, error) {

	query := "SELECT company_id FROM companies WHERE project_id = $1"
	var company_id string

	tx := proj.DB.Raw(query, projectID).Scan(&company_id)
	if tx.Error != nil {
		return false, "", tx.Error
	}

	if tx.RowsAffected == 0 {
		return false, "", nil
	}

	return true, company_id, nil
}

func (proj *ProjectAdapter) MemberState(userID, projectID string) (string, error) {

	query := "SELECT s.status FROM member_statuses s INNER JOIN members m ON s.id = m.status_id AND m.member_id = $1 AND m.project_id = $2"
	var res string

	tx := proj.DB.Raw(query, userID, projectID).Scan(&res)
	if tx.Error != nil {
		return "", tx.Error
	}

	if tx.RowsAffected == 0 {
		return "", errors.New("the user doesnt exist")
	}

	return res, nil
}

func (project *ProjectAdapter) GetLiveProjectsofCompany(compID string) ([]entities.GetLiveProjectsUsecase, error) {

	/* query := "SELECT c.project_id,c.project_username,c.description,COUNT(*) FROM credentials c INNER JOIN companies es ON es.project_id = c.project_id AND es.company_id = $1 RIGHT JOIN members m ON m.project_id = c.project_id INNER JOIN member_statuses ms ON ms.id = m.status_id AND ms.status = 'ACCEPTED' GROUP BY ////(c.project_id,c.project_username,c.description)" */
	query := "SELECT c.project_id,c.project_username,c.description FROM credentials c INNER JOIN companies es ON es.project_id = c.project_id AND es.company_id = $1"
	var res []entities.GetLiveProjectsUsecase

	if err := project.DB.Raw(query, compID).Scan(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func (project *ProjectAdapter) GetCompletedMembers(projectID string, users []string) ([]entities.GetCompletedMemebersAdapter, error) {

	var res []entities.GetCompletedMemebersAdapter
	if err := project.DB.Model(&entities.TaskAssignations{}).Where("project_id = ? AND user_id IN ?", projectID, users).Pluck("stages,is_verified", &res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func (project *ProjectAdapter) RaiseIssue(req entities.Issues) error {

	query := "INSERT INTO issues (project_id,user_id,description,raiser_id) VALUES($1,$2,$3,$4)"
	if err := project.DB.Exec(query, req.ProjectID, req.UserID, req.Description, req.RaiserID).Error; err != nil {
		return err
	}

	return nil
}

func (proj *ProjectAdapter) GetIssuesofMember(projecID, userID string) (entities.Issues, error) {

	var res entities.Issues
	query := "SELECT * FROM issues WHERE project_id = $1 AND user_id = $2"

	if err := proj.DB.Raw(query, projecID, userID).Scan(&res).Error; err != nil {
		return entities.Issues{}, err
	}

	return res, nil
}

func (proj *ProjectAdapter) GetIssuesofProject(projectID string) ([]entities.Issues, error) {

	var res []entities.Issues
	query := "SELECT * FROM issues WHERE project_id = $1"

	if err := proj.DB.Raw(query, projectID).Scan(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func (proj *ProjectAdapter) RateTask(req entities.Ratings) error {

	query := "INSERT INTO ratings (project_id,user_id,rating,feedback) VALUES($1,$2,$3,$4)"

	if err := proj.DB.Exec(query, req.ProjectID, req.UserID, req.Rating, req.Feedback).Error; err != nil {
		return err
	}

	return nil
}

func (proj *ProjectAdapter) GetRating(projectID, userID string) (entities.Ratings, error) {

	query := "SELECT * FROM ratings WHERE project_id = $1 AND user_id = $2"
	var res entities.Ratings

	if err := proj.DB.Raw(query, projectID, userID).Scan(&res).Error; err != nil {
		return entities.Ratings{}, err
	}

	return res, nil
}

func (proj *ProjectAdapter) AskExtension(req entities.Extensions) error {

	query := "INSERT INTO extensions (project_id,user_id,extend_to,description) VALUES($1,$2,$3,$4)"
	if err := proj.DB.Exec(query, req.ProjectID, req.UserID, req.ExtendTo, req.Description).Error; err != nil {
		return err
	}

	return nil
}

func (proj *ProjectAdapter) GetExtensionRequestsinaProject(projectID string) ([]entities.Extensions, error) {

	var res []entities.Extensions
	query := "SELECT * FROM extensions WHERE project_id = $1"

	if err := proj.DB.Raw(query, projectID).Scan(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func (proj *ProjectAdapter) ApproveExtensionRequest(id uint, isGranted bool) error {

	query := "UPDATE extensions SET is_accepted = $1 WHERE id = $2"

	if err := proj.DB.Exec(query, isGranted, id).Error; err != nil {
		return err
	}

	return nil
}

func (proj *ProjectAdapter) VerifyTaskCompletion(projectID, memberID string, verified bool) error {

	query := "UPDATE task_assignations SET is_verified = true WHERE project_id = $1 AND user_id = $2"

	if err := proj.DB.Exec(query, projectID, memberID).Error; err != nil {
		return err
	}

	return nil
}

func (proj *ProjectAdapter) GetVerifiedTasks(projectID string) ([]entities.VerifiedTasksUsecase, error) {

	var res []entities.VerifiedTasksUsecase
	query := "SELECT t.user_id,r.rating,r.feedback FROM task_assignations t INNER JOIN ratings r ON r.project_id = $1 AND r.user_id = t.user_id WHERE t.project_id = $1 AND is_verified = true"

	if err := proj.DB.Raw(query, projectID).Scan(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func (proj *ProjectAdapter) DropProject(projID string) error {

	if err := proj.DB.Unscoped().Delete(&entities.Credentials{}, "project_id = ?", projID).Error; err != nil {
		return err
	}

	return nil
}

func (proj *ProjectAdapter) EditProject(req entities.Credentials) error {

	if err := proj.DB.Model(&entities.Credentials{}).Where("project_id = ?", req.ProjectID).Updates(req).Error; err != nil {
		return err
	}

	return nil
}

func (proj *ProjectAdapter) EditMember(req entities.Members) error {

	if err := proj.DB.Model(&entities.Members{}).Where("project_id = $1 AND member_id = $2", req.ProjectID, req.MemberID).Updates(req).Error; err != nil {
		return err
	}

	return nil
}

func (proj *ProjectAdapter) EditFeedback(req entities.Ratings) error {

	if err := proj.DB.Model(&entities.Ratings{}).Where("project_id = $1 AND user_id = $2", req.ProjectID, req.UserID).Updates(req).Error; err != nil {
		return err
	}

	return nil
}

func (proj *ProjectAdapter) DeleteFeedback(projectID, memberID string) error {

	if err := proj.DB.Unscoped().Delete(&entities.Ratings{}, "project_id = $1 AND user_id", projectID, memberID).Error; err != nil {
		return err
	}

	return nil
}

func (proj *ProjectAdapter) GetCountMembers(projectID string) (uint, error) {

	query := "SELECT COUNT(*) FROM members m INNER JOIN member_statuses ms ON m.status_id = ms.id AND ms.status = 'ACCEPTED' WHERE project_id = $1 GROUP BY project_id"
	var count uint

	if err := proj.DB.Raw(query, projectID).Scan(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (proj *ProjectAdapter) TerminateProjectMembers(userID, projID string) error {

	query := "UPDATE members SET status_id = (SELECT id FROM member_statuses WHERE status = 'TERMINATED') WHERE member_id = $1 AND project_id = $2"

	if err := proj.DB.Exec(query, userID, projID).Error; err != nil {
		return err
	}

	return nil
}
