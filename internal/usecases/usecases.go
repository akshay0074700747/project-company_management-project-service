package usecases

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/akshay0074700747/project-company_management-project-service/entities"
	"github.com/akshay0074700747/project-company_management-project-service/helpers"
	"github.com/akshay0074700747/project-company_management-project-service/internal/adapters"
	"github.com/minio/minio-go/v7"
)

type ProjectUseCases struct {
	Adapter adapters.ProjectAdapterInterfaces
}

func NewProjectUseCases(adapter adapters.ProjectAdapterInterfaces) *ProjectUseCases {
	return &ProjectUseCases{
		Adapter: adapter,
	}
}

func (project *ProjectUseCases) CreateProject(req entities.Credentials, compID, ownerID string) (entities.Credentials, error) {

	if req.ProjectUsername == "" {
		return entities.Credentials{}, errors.New("the project username cannot be empty")
	}

	if req.Name == "" {
		return entities.Credentials{}, errors.New("the project rname cannot be empty")
	}

	isExisting, err := project.Adapter.IsProjectUsernameExists(req.ProjectUsername)
	if err != nil {
		helpers.PrintErr(err, "error occured at IsProjectUsernameExists adapter")
	}

	if isExisting {
		return entities.Credentials{}, errors.New("the username is already taken")
	}

	req.ProjectID = helpers.GenUuid()

	res, err := project.Adapter.CreateProject(req, compID)
	if err != nil {
		helpers.PrintErr(err, "error occured at CreateProject adapter")
		return entities.Credentials{}, err
	}
	if err = project.Adapter.AddOwner(res.ProjectID, ownerID); err != nil {
		helpers.PrintErr(err, "error at AddOwner adapter")
		return entities.Credentials{}, err
	}

	return res, nil

}

func (project *ProjectUseCases) Addmembers(req entities.Members) error {

	isExists, err := project.Adapter.IsMemberExists(req.MemberID, req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "error occures at IsMemberExists adapter")
		return errors.New("there has been an error , please try again later")
	}

	if isExists {
		return errors.New("the member already exists")
	}

	if err = project.Adapter.AddMember(req); err != nil {
		helpers.PrintErr(err, "error occures at AddMember adapter")
		return errors.New("there has been an error , please try again later")
	}
	return nil
}

func (project *ProjectUseCases) AcceptInvitation(req entities.Members, accepted bool) error {

	if accepted {
		if err := project.Adapter.AcceptInvitation(req); err != nil {
			helpers.PrintErr(err, "error occured at AcceptInvitation adapter")
			return err
		}
	} else {
		if err := project.Adapter.DenyInvitation(req); err != nil {
			helpers.PrintErr(err, "error occured at DenyInvitation adapter")
			return err
		}
	}

	return nil
}

func (project *ProjectUseCases) GetProjectInvites(memID string) ([]entities.ProjectInviteUsecase, error) {

	res, err := project.Adapter.GetProjectInvites(memID)
	if err != nil {
		helpers.PrintErr(err, "error at GetProjectInvites adapter")
		return nil, err
	}

	return res, nil
}

func (project *ProjectUseCases) GetProjectDetails(projectID string, projectUsername string) (entities.ProjectDetailsUsecase, error) {

	cred, err := project.Adapter.GetProjectDetails(projectID)
	if err != nil {
		helpers.PrintErr(err, "error at GetProjectDetails adapter")
		return entities.ProjectDetailsUsecase{}, err
	}
	mem, err := project.Adapter.GetNoofMembers(projectID)
	if err != nil {
		helpers.PrintErr(err, "error at GetNoofMembers adapter")
		return entities.ProjectDetailsUsecase{}, err
	}

	return entities.ProjectDetailsUsecase{
		ProjectID:       cred.ProjectID,
		ProjectUsername: cred.ProjectUsername,
		Aim:             cred.Aim,
		Members:         mem,
	}, nil
}

func (project *ProjectUseCases) GetProjectMembers(projectID string) ([]entities.GetProjectMembersUsecase, error) {

	res, err := project.Adapter.GetProjectMembers(projectID)
	if err != nil {
		helpers.PrintErr(err, "error at GetProjectMembers adapter")
		return nil, err
	}

	return res, nil
}

func (project *ProjectUseCases) AddMemberStatueses(status string) error {

	if err := project.Adapter.AddMemberStatueses(status); err != nil {
		helpers.PrintErr(err, "error occured at AddMemberStatueses")
		return err
	}

	return nil
}

func (project *ProjectUseCases) AssignTasks(req entities.TaskDta) error {

	objectName := helpers.GenUuid()

	newReader := bytes.NewReader(req.File)
	err := project.Adapter.InsertTasktoMinio(context.TODO(), objectName, newReader, newReader.Size(), minio.PutObjectOptions{
		UserMetadata: map[string]string{
			"userID":    req.UserID,
			"projectID": req.ProjectID,
		},
	})
	if err != nil {
		helpers.PrintErr(err, "error happened at InsertTasktoMinio adapter")
		return err
	}

	if err = project.Adapter.InsertTaskDetails(entities.TaskAssignations{
		UserID:      req.UserID,
		ProjectID:   req.ProjectID,
		Task:        req.Task,
		Description: req.Description,
		ObjectName:  objectName,
		Stages:      req.Stages,
		Deadline:    req.Deadline,
	}); err != nil {
		helpers.PrintErr(err, "error happened at InsertTasktoMinio adapter")
		return err
	}

	return nil
}

func (project *ProjectUseCases) GetTaskDetails(userId, projectID string) (entities.TaskAssignations, error) {

	res, err := project.Adapter.GetTaskDetails(userId, projectID)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetTaskDetails adapter")
		return entities.TaskAssignations{}, err
	}

	return res, nil
}

func (project *ProjectUseCases) DownloadTask(objectName string) ([]byte, error) {

	file, err := project.Adapter.GetTaskFromMinio(context.TODO(), objectName, minio.GetObjectOptions{})
	if err != nil {
		helpers.PrintErr(err, "error happened at GetTaskFromMinio adapter")
		return nil, err
	}

	return file, nil
}

func (project *ProjectUseCases) InsertStatuses(req entities.TaskStatuses) error {

	if err := project.Adapter.InsertStatuses(req); err != nil {
		helpers.PrintErr(err, "errror happened at InsertStatuses")
		return err
	}

	return nil
}

func (project *ProjectUseCases) GetProgressofMember(req entities.UserProgressUsecaseRes) (entities.UserProgressUsecaseRes, error) {

	taskDtl, err := project.Adapter.GetTaskDetails(req.MemberID, req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "errror happened at GetTaskDetails adapter")
		return entities.UserProgressUsecaseRes{}, err
	}

	progress := (req.TasksCompleted / uint32(taskDtl.Stages)) * 100

	req.Progress = (strconv.Itoa(int(progress)) + "%")
	req.TaskDeadline = taskDtl.Deadline.String()
	req.TasksLeft = uint32(taskDtl.Stages) - req.TasksCompleted

	return req, nil

}

func (project *ProjectUseCases) GetRoleID(usrID string) (uint, error) {

	role, err := project.Adapter.GetRoleID(usrID)
	if err != nil {
		helpers.PrintErr(err, "errror happened at GetRoleID adapter")
		return 0, err
	}

	return role, nil
}

func (project *ProjectUseCases) LogintoProject(usrName, memberID string) (entities.Members, error) {

	projectId, err := project.Adapter.GetIDfromName(usrName)
	if err != nil {
		helpers.PrintErr(err, "errror happened at GetIDfromName adapter")
		return entities.Members{}, err
	}
	res, err := project.Adapter.LogintoProject(projectId, memberID)
	if err != nil {
		helpers.PrintErr(err, "errror happened at LogintoProject adapter")
		return entities.Members{}, err
	}

	res.ProjectID = projectId

	return res, nil
}

func (project *ProjectUseCases) GetProgressofMembers(req entities.ListofUserProgress, projectID string) (entities.ListofUserProgress, []string, []uint, error) {

	var users []string
	for _, v := range req.UserAndProgress {
		users = append(users, v.UserID)
	}
	res, err := project.Adapter.GetStagesandDeadline(users, projectID)
	if err != nil {
		helpers.PrintErr(err, "errror happened at LogintoProject adapter")
		return entities.ListofUserProgress{}, nil, nil, err
	}

	var progresses []int
	for i := range req.UserAndProgress {
		req.UserAndProgress[i].TaskDeadline = res[i].Deadline.String()
		progress := (float32(req.UserAndProgress[i].Stages) / float32(res[i].Stages)) * 100
		req.UserAndProgress[i].Progress = (strconv.Itoa((int(progress))) + "%")
		progg := int(progress)
		progg = progg - (progg % 10)
		progresses = append(progresses, progg)
	}

	statuses, err := project.Adapter.GetTaskStatuses(progresses)
	if err != nil {
		helpers.PrintErr(err, "errror happened at LogintoProject adapter")
		return entities.ListofUserProgress{}, nil, nil, err
	}

	roleIds, err := project.Adapter.GetListofRoleIds(users, projectID)
	if err != nil {
		helpers.PrintErr(err, "errror happened at GetListofRoleIds adapter")
		return entities.ListofUserProgress{}, nil, nil, err
	}

	return req, statuses, roleIds, nil
}

func (project *ProjectUseCases) InsertNonTechnicalTasks(req entities.NonTechnicalTaskDetials) error {

	if err := project.Adapter.InsertNonTechnicalTasks(req); err != nil {
		helpers.PrintErr(err, "errror happened at InsertNonTechnicalTasks adapter")
		return err
	}

	return nil
}

func (project *ProjectUseCases) GetProjectProgress(projectID string, req entities.ListofUserProgress) (entities.GetProjectProgressUsecase, error) {

	var result entities.GetProjectProgressUsecase
	result.ProjectID = projectID
	deadline, err := project.Adapter.GetProjectDeadline(projectID)
	if err != nil {
		helpers.PrintErr(err, "errror happened at GetProjectDeadline adapter")
		return entities.GetProjectProgressUsecase{}, err
	}

	members, err := project.Adapter.GetCountofLivemembers(projectID)
	if err != nil {
		helpers.PrintErr(err, "errror happened at GetCountofLivemembers adapter")
		return entities.GetProjectProgressUsecase{}, err
	}
	var fullCompleted, lessCompleted int
	var sum float32
	for _, v := range req.UserAndProgress {
		ttStages, err := project.Adapter.GetStagesofProgress(projectID, v.UserID)
		if err != nil {
			helpers.PrintErr(err, "errror happened at GetStagesofProgress adapter")
			return entities.GetProjectProgressUsecase{}, err
		}
		progress := (float32(v.Stages) / float32(ttStages)) * 100
		if progress == 100 {
			fullCompleted++
		} else if progress < 30 {
			lessCompleted++
		}
		sum += progress
	}

	result.Deadline = deadline.String()
	result.LiveMembers = uint32(members)
	result.TaskCompletedMembers = uint32(fullCompleted)
	result.TaskCriticalMembers = uint32(lessCompleted)
	result.Progress = (strconv.Itoa((int(sum / float32(len(req.UserAndProgress))))) + "%")

	return result, nil
}

func (project *ProjectUseCases) IsOwner(user_id, project_id string) (bool, error) {

	res, err := project.Adapter.IsOwner(user_id, project_id)
	if err != nil {
		helpers.PrintErr(err, "error at IsOwner adapter")
		return false, err
	}

	if res {
		return true, nil
	}
	return false, nil
}

func (proj *ProjectUseCases) IsCompanyBased(projID string) (bool, string, error) {

	isCompanybased, companyID, err := proj.Adapter.IsCompanyBased(projID)
	if err != nil {
		return false, "", err
	}

	if companyID == "" || !isCompanybased {
		return false, "", nil
	}

	return true, companyID, nil
}

func (proj *ProjectUseCases) IsMemberAccepted(userID, projectID string) error {

	state, err := proj.Adapter.MemberState(userID, projectID)
	if err != nil {
		helpers.PrintErr(err, "error happened at MemberState adapter")
		return err
	}

	if state != "ACCEPTED" {
		return errors.New("the member is not accepted the project invitation")
	}

	return nil
}

func (project *ProjectUseCases) GetLiveProjectsofCompany(compID string) ([]entities.GetLiveProjectsUsecase, error) {

	res, err := project.Adapter.GetLiveProjectsofCompany(compID)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetLiveProjectsofCompany adapter")
		return nil, err
	}

	return res, nil
}

func (project *ProjectUseCases) GetCompletedMembers(projectID string, req entities.ListofUserProgress, isFromCompleted bool) ([]entities.GetCompletedMemebersUsecase, error) {

	var userIds []string
	var result []entities.GetCompletedMemebersUsecase

	for _, v := range req.UserAndProgress {
		userIds = append(userIds, v.UserID)
	}

	res, err := project.Adapter.GetCompletedMembers(projectID, userIds)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetCompletedMembers adapter")
		return nil, err
	}

	if isFromCompleted {

		fmt.Println(len(res), "hvkjghj")
		fmt.Println(len(req.UserAndProgress), "dkhvlcjewoij")

		for i, v := range res {
			if i >= len(req.UserAndProgress) {
				break
			}

			progress := (req.UserAndProgress[i].Stages / v.Stages)
			if progress == 1 {
				result = append(result, entities.GetCompletedMemebersUsecase{
					UserID:     req.UserAndProgress[i].UserID,
					IsVerified: v.IsVerified,
				})
			}

		}

	} else {

		for i, v := range res {

			progress := (float32(req.UserAndProgress[i].Stages) / float32(v.Stages))
			if progress <= 0.3 {
				result = append(result, entities.GetCompletedMemebersUsecase{
					UserID:     req.UserAndProgress[i].UserID,
					IsVerified: v.IsVerified,
				})
			}

		}

	}

	return result, nil
}

func (project *ProjectUseCases) RaiseIssue(req entities.Issues) error {

	if err := project.Adapter.RaiseIssue(req); err != nil {
		helpers.PrintErr(err, "error happened at RaiseIssue adapter")
		return err
	}

	return nil
}

func (proj *ProjectUseCases) GetIssuesofMember(projectID, userID string) (entities.Issues, error) {

	res, err := proj.Adapter.GetIssuesofMember(projectID, userID)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetIssuesofMember adapter")
		return entities.Issues{}, err
	}

	return res, nil
}

func (proj *ProjectUseCases) GetIssuesofProject(projID string) ([]entities.Issues, error) {

	res, err := proj.Adapter.GetIssuesofProject(projID)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetIssuesofProject adpaters")
		return nil, err
	}

	return res, nil
}

func (proj *ProjectUseCases) RateTask(req entities.Ratings) error {

	if err := proj.Adapter.RateTask(req); err != nil {
		helpers.PrintErr(err, "error happened at RateTask adapter")
		return err
	}

	return nil
}

func (proj *ProjectUseCases) GetRating(projectID, userID string) (entities.Ratings, error) {

	res, err := proj.Adapter.GetRating(projectID, userID)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetRating adapter")
		return entities.Ratings{}, err
	}

	return res, nil
}

func (proj *ProjectUseCases) AskExtension(req entities.Extensions) error {

	if err := proj.Adapter.AskExtension(req); err != nil {
		helpers.PrintErr(err, "error happened at AskExtension adapter")
		return err
	}

	return nil
}

func (proj *ProjectUseCases) GetExtensionRequestsinaProject(projectID string) ([]entities.Extensions, error) {

	res, err := proj.Adapter.GetExtensionRequestsinaProject(projectID)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetExtensionRequestsinaProject adapter")
		return nil, err
	}

	return res, nil
}

func (proj *ProjectUseCases) ApproveExtensionRequest(id uint, isAccepted bool) error {

	if err := proj.Adapter.ApproveExtensionRequest(id, isAccepted); err != nil {
		helpers.PrintErr(err, "eroror happened at ApproveExtensionRequest adapter")
		return err
	}

	return nil
}

func (proj *ProjectUseCases) VerifyTaskCompletion(projectID, userID string, verified bool) error {

	if err := proj.Adapter.VerifyTaskCompletion(projectID, userID, verified); err != nil {
		helpers.PrintErr(err, "errro happened at VerifyTaskCompletion adapter")
		return err
	}

	return nil
}

func (proj *ProjectUseCases) GetVerifiedTasks(projectID string) ([]entities.VerifiedTasksUsecase, error) {

	res, err := proj.Adapter.GetVerifiedTasks(projectID)
	if err != nil {
		helpers.PrintErr(err, "error happerned at GetVerifiedTasks adapter")
		return nil, err
	}

	return res, nil
}

func (proj *ProjectUseCases) DropProject(projectID string) error {

	if er := proj.Adapter.DropProject(projectID); er != nil {
		helpers.PrintErr(er, "error happened at DropProject adapter")
		return er
	}

	return nil
}

func (proj *ProjectUseCases) EditProject(req entities.Credentials) error {

	if err := proj.Adapter.EditProject(req); err != nil {
		helpers.PrintErr(err, "error happened at EditProject adapter")
		return err
	}

	return nil
}

func (proj *ProjectUseCases) EditMember(req entities.Members) error {

	if err := proj.Adapter.EditMember(req); err != nil {
		helpers.PrintErr(err, "error happened at EditMember adapter")
		return err
	}

	return nil
}

func (proj *ProjectUseCases) EditFeedback(req entities.Ratings) error {
	if err := proj.Adapter.EditFeedback(req); err != nil {
		helpers.PrintErr(err, "error happened at EditFeedback adapter")
		return err
	}

	return nil
}

func (proj *ProjectUseCases) DeleteFeedback(projectID, userID string) error {

	if err := proj.Adapter.DeleteFeedback(projectID, userID); err != nil {
		helpers.PrintErr(err, "error happened at DeleteFeedback adapter")
		return err
	}

	return nil
}

func (proj *ProjectUseCases) GetCountMembers(projectID string) (uint, error) {

	res, err := proj.Adapter.GetCountMembers(projectID)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetCountMembers adapter")
		return 0, err
	}

	return res, err
}

func (proj *ProjectUseCases) IsMemberExists(userID, projectID string) (bool, error) {

	res, err := proj.Adapter.IsMemberExists(userID, projectID)
	if err != nil {
		helpers.PrintErr(err, "error happend at IsMemberExists adapter")
		return false, err
	}

	return res, nil
}

func (proj *ProjectUseCases) TerminateProjectMembers(userID, projID string) error {

	if err := proj.Adapter.TerminateProjectMembers(userID, projID); err != nil {
		helpers.PrintErr(err, "error happenda t TerminateProjectMembers adapter")
		return err
	}

	return nil
}
