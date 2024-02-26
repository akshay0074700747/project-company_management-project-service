package usecases

import (
	"bytes"
	"context"
	"errors"
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

	if !req.IsCompanybased {
		if err = project.Adapter.AddOwner(res.ProjectID, ownerID); err != nil {
			helpers.PrintErr(err, "error at AddOwner adapter")
			return entities.Credentials{}, err
		}
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

	if projectID != "" {

	} else {

	}

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
		ProjectID:   req.UserID,
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
		progress := (req.UserAndProgress[i].Stages / res[i].Stages) * 100
		req.UserAndProgress[i].Progress = (strconv.Itoa((progress)) + "%")
		progress = progress - (progress % 10)
		progresses = append(progresses, progress)
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
	var fullCompleted, lessCompleted, sum int
	for _, v := range req.UserAndProgress {
		ttStages, err := project.Adapter.GetStagesofProgress(projectID, v.UserID)
		if err != nil {
			helpers.PrintErr(err, "errror happened at GetStagesofProgress adapter")
			return entities.GetProjectProgressUsecase{}, err
		}
		progress := (v.Stages / ttStages) * 100
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
	result.Progress = (strconv.Itoa((int(sum / len(req.UserAndProgress)))) + "%")

	return result, nil
}
