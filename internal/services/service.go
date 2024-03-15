package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/akshay0074700747/project-company_management-project-service/entities"
	"github.com/akshay0074700747/project-company_management-project-service/helpers"
	"github.com/akshay0074700747/project-company_management-project-service/internal/usecases"
	"github.com/akshay0074700747/projectandCompany_management_protofiles/pb/companypb"
	"github.com/akshay0074700747/projectandCompany_management_protofiles/pb/projectpb"
	"github.com/akshay0074700747/projectandCompany_management_protofiles/pb/userpb"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis/v8"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProjectServiceServer struct {
	Usecase     usecases.ProjectUsecaseInterfaces
	UserConn    userpb.UserServiceClient
	CompanyConn companypb.CompanyServiceClient
	Producer    *kafka.Producer
	Topic       string
	Cache       *redis.Client
	projectpb.UnimplementedProjectServiceServer
}

func NewProjectServiceServer(usecase usecases.ProjectUsecaseInterfaces, usraddr, compaddr, topic string, producer *kafka.Producer) *ProjectServiceServer {
	userRes, _ := helpers.DialGrpc(usraddr)
	compRes, _ := helpers.DialGrpc(compaddr)
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", 
		Password: "",               
		DB:       0,                
	})
	return &ProjectServiceServer{
		Usecase:     usecase,
		UserConn:    userpb.NewUserServiceClient(userRes),
		CompanyConn: companypb.NewCompanyServiceClient(compRes),
		Topic:       topic,
		Producer:    producer,
		Cache:       rdb,
	}
}

func (project *ProjectServiceServer) CreateProject(tx context.Context, req *projectpb.CreateProjectReq) (*projectpb.CreateProjectRes, error) {

	res, err := project.Usecase.CreateProject(entities.Credentials{
		ProjectUsername: req.ProjectUsername,
		Name:            req.Name,
		Aim:             req.Aim,
		Description:     req.Description,
		IsCompanybased:  req.IsCompanybased,
		IsPublic:        req.IsPublic,
	}, req.ComapanyId, req.OwnerID)

	if err != nil {
		helpers.PrintErr(err, "error occured at CreateProject usecase")
		return nil, err
	}

	return &projectpb.CreateProjectRes{
		ProjectID:       res.ProjectID,
		ProjectUsername: res.ProjectUsername,
		Name:            res.Name,
		Aim:             res.Aim,
		Description:     res.Description,
		IsCompanybased:  res.IsCompanybased,
		ComapanyID:      req.ComapanyId,
	}, nil

}

func (project *ProjectServiceServer) AcceptProjectInvite(ctx context.Context, req *projectpb.AcceptProjectInviteReq) (*emptypb.Empty, error) {

	if err := project.Usecase.AcceptInvitation(entities.Members{
		MemberID:  req.UserID,
		ProjectID: req.ProjectID,
	}, req.Accept); err != nil {
		helpers.PrintErr(err, "error occured at AcceptInvitation usecase...")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) AddMembers(ctx context.Context, req *projectpb.AddMemberReq) (*emptypb.Empty, error) {

	isCompanybased, companyID, err := project.Usecase.IsCompanyBased(req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "error at IsCompanyBased usecase")
		return nil, err
	}
	res, err := project.UserConn.GetByEmail(ctx, &userpb.GetByEmailReq{
		Email: req.Email,
	})
	if err != nil {
		return nil, errors.New("cannot connect to user-service now please try again later")
	}

	if isCompanybased {
		res, err := project.CompanyConn.IsEmployeeExists(ctx, &companypb.IsEmployeeExistsReq{
			CompanyID:  companyID,
			EmployeeID: res.UserID,
		})
		if err != nil {
			return nil, err
		}
		if !res.Exists {
			return nil, errors.New("the employee doesnt exist in the company")
		}
	}

	count, err := project.Usecase.GetCountMembers(req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "error at GetCountMembers usecase")
		return nil, err
	}

	if count > 10 {

		url := fmt.Sprintf("http://localhost:50007/transaction/project?assetID=%s", req.ProjectID)
		resStages, err := http.Get(url)
		if err != nil {
			helpers.PrintErr(err, "errro happened at calling http method")
			return nil, err
		}

		var ress entities.Responce
		if err := json.NewDecoder(resStages.Body).Decode(&ress); err != nil {
			helpers.PrintErr(err, "errro happened at decoding the json")
			return nil, err
		}

		if !ress.Data.(bool) {
			helpers.PrintErr(err, "errro happened at decoding the json")
			return nil, errors.New("you need to purchase premium for adding more than 10 members")
		}
	}

	if err := project.Usecase.Addmembers(entities.Members{
		MemberID:     res.UserID,
		ProjectID:    req.ProjectID,
		PermissionID: uint(req.PermissionID),
		RoleID:       uint(req.RoleID),
	}); err != nil {
		helpers.PrintErr(err, "error at Addmembers usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) ProjectInvites(req *projectpb.ProjectInvitesReq, stream projectpb.ProjectService_ProjectInvitesServer) error {
	res, err := project.Usecase.GetProjectInvites(req.MemberID)
	if err != nil {
		helpers.PrintErr(err, "error occured at GetProjectInvites usecase")
		return err
	}
	for _, v := range res {
		if err = stream.Send(&projectpb.ProjectInvitesRes{
			ProjectID: v.ProjectID,
			Members:   uint32(v.AcceptedMembers),
		}); err != nil {
			helpers.PrintErr(err, "error at sending stream")
			return errors.New("cannot process you request now , please try again later")
		}
	}

	return nil
}

func (project *ProjectServiceServer) GetProjectDetailes(ctx context.Context, req *projectpb.GetProjectReq) (*projectpb.GetProjectDetailesRes, error) {

	res, err := project.Usecase.GetProjectDetails(req.ProjectID, req.ProjectUsername)
	if err != nil {
		helpers.PrintErr(err, "error at GetProjectDetails usecase")
		return nil, errors.New("cannot connect to the project service now")
	}

	return &projectpb.GetProjectDetailesRes{
		ProjectID:       res.ProjectID,
		ProjectUsername: res.ProjectUsername,
		Members:         uint32(res.Members),
		Aim:             res.Aim,
	}, nil
}

func (project *ProjectServiceServer) GetProjectMembers(req *projectpb.GetProjectReq, stream projectpb.ProjectService_GetProjectMembersServer) error {

	res, err := project.Usecase.GetProjectMembers(req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "error at GetProjectMembers usecase")
		return err
	}

	streaam, err := project.UserConn.GetStreamofUserDetails(context.TODO())
	if err != nil {
		helpers.PrintErr(err, "error at GetStreamofUserDetails usecase")
		return err
	}

	for _, v := range res {

		if err := streaam.Send(&userpb.GetUserDetailsReq{
			UserID: v.UserID,
			RoleID: uint32(v.RoleID),
		}); err != nil {
			helpers.PrintErr(err, "error at sending to stream")
			return err
		}

		details, err := streaam.Recv()
		if err != nil {
			helpers.PrintErr(err, "error at recieving to stream")
			return err
		}

		if err := stream.Send(&projectpb.GetProjectMembersRes{
			UserID:       details.UserID,
			Email:        details.Email,
			Name:         details.Name,
			Role:         details.Role,
			PermissionID: uint32(v.PermissionID),
		}); err != nil {
			helpers.PrintErr(err, "error at sending stream")
			return err
		}
	}

	return nil
}

func (project *ProjectServiceServer) LogintoProject(ctx context.Context, req *projectpb.LogintoProjectReq) (*projectpb.LogintoProjectRes, error) {

	res, err := project.Usecase.LogintoProject(req.ProjectUsername, req.MemberID)
	if err != nil {
		helpers.PrintErr(err, "error at sending stream")
		return nil, err
	}

	fmt.Println(res, "----result")
	fmt.Println(req.MemberID, "----request id")

	isOwnerbool, err := project.Usecase.IsOwner(req.MemberID, res.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "error at IsOwner usecase")
		return nil, err
	}

	if isOwnerbool {
		return &projectpb.LogintoProjectRes{
			ProjectID:  res.ProjectID,
			Permission: "ROOT",
			Role:       "OWNER",
		}, nil
	}

	if res.ProjectID == "" || res.PermissionID == 0 || res.RoleID == 0 {
		return nil, errors.New("the projectuseranme doesnt exist or user is not part of the project")
	}

	compRes, err := project.CompanyConn.GetPermission(ctx, &companypb.GetPermisssionReq{
		ID: uint32(res.PermissionID),
	})
	if err != nil {
		helpers.PrintErr(err, "error at communicating with the company service")
		return nil, err
	}

	roleRes, err := project.UserConn.GetRole(ctx, &userpb.GetRoleReq{
		ID: uint32(res.RoleID),
	})
	if err != nil {
		helpers.PrintErr(err, "error at communicating with the company service")
		return nil, err
	}

	return &projectpb.LogintoProjectRes{
		ProjectID:  res.ProjectID,
		Permission: compRes.Permission,
		Role:       roleRes.Role,
	}, nil
}

func (project *ProjectServiceServer) AddMemberStatus(ctx context.Context, req *projectpb.MemberStatusReq) (*emptypb.Empty, error) {

	if err := project.Usecase.AddMemberStatueses(req.Status); err != nil {
		helpers.PrintErr(err, "error happed at AddMemberStatueses")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) DownloadTask(ctx context.Context, req *projectpb.DownloadTaskReq) (*projectpb.DownloadTaskRes, error) {

	task, err := project.Usecase.GetTaskDetails(req.UserID, req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "error happed at GetTaskDetails usecase")
		return nil, err
	}

	if task.ObjectName != req.TaskFileID {
		helpers.PrintErr(err, "the tASKFILEID and objectNmae is no same")
		return nil, errors.New("the given task file id doesnt exists")
	}

	res, err := project.Usecase.DownloadTask(task.ObjectName)
	if err != nil {
		helpers.PrintErr(err, "error happed at DownloadTask usecase")
		return nil, err
	}

	return &projectpb.DownloadTaskRes{
		File: res,
	}, nil
}

func (project *ProjectServiceServer) GetAssignedTask(ctx context.Context, req *projectpb.GetAssignedTaskReq) (*projectpb.GetAssignedTaskRes, error) {

	task, err := project.Usecase.GetTaskDetails(req.UserID, req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "error happed at GetTaskDetails usecase")
		return nil, err
	}

	return &projectpb.GetAssignedTaskRes{
		Task:        task.Task,
		Description: task.Description,
		TaskFileID:  task.ObjectName,
		Stages:      uint32(task.Stages),
		Deadline:    timestamppb.New(task.Deadline),
	}, nil
}

func (project *ProjectServiceServer) AddTaskStatuses(ctx context.Context, req *projectpb.AddTaskStatusesReq) (*emptypb.Empty, error) {

	if err := project.Usecase.InsertStatuses(entities.TaskStatuses{
		Status: req.Status,
		Stat:   int(req.Below),
	}); err != nil {
		helpers.PrintErr(err, "error happened at InsertStatuses usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) GetProgressofMember(ctx context.Context, req *projectpb.GetProgressofMemberReq) (*projectpb.GetProgressofMemberRes, error) {

	url := fmt.Sprintf("http://localhost:50005/project/task/stages?userID=%s&&projectID=%s", req.MemberID, req.ProjectID)
	resStages, err := http.Get(url)
	if err != nil {
		helpers.PrintErr(err, "errro happened at calling http method")
		return nil, err
	}

	var res entities.StageRes
	if err := json.NewDecoder(resStages.Body).Decode(&res); err != nil {
		helpers.PrintErr(err, "errro happened at decoding the json")
		return nil, err
	}

	roleID, err := project.Usecase.GetRoleID(req.MemberID)
	if err != nil {
		helpers.PrintErr(err, "errro happened at GetRoleID usecase")
		return nil, err
	}

	usrDetails, err := project.UserConn.GetUserDetails(ctx, &userpb.GetUserDetailsReq{
		UserID: req.MemberID,
		RoleID: uint32(roleID),
	})
	if err != nil {
		helpers.PrintErr(err, "errro happened at GetUserDetails")
		return nil, err
	}

	result, err := project.Usecase.GetProgressofMember(entities.UserProgressUsecaseRes{
		MemberID:       req.MemberID,
		ProjectID:      req.ProjectID,
		TasksCompleted: uint32(res.Stages),
	})
	if err != nil {
		helpers.PrintErr(err, "errro happened at GetProgressofMember usecase")
		return nil, err
	}

	var ress projectpb.GetProgressofMemberRes

	for _, v := range res.Details {
		ress.Details = append(ress.Details, &projectpb.TaskDetails{
			Key:              v.Key,
			Description:      v.Description,
			FileSnapshotName: v.Filename,
		})
	}
	ress.Email = usrDetails.Email
	ress.Name = usrDetails.Name
	ress.Role = usrDetails.Role
	ress.MemberID = result.MemberID
	ress.Progress = result.Progress
	ress.TaskDeadline = result.TaskDeadline
	ress.TasksCompleted = result.TasksCompleted
	ress.TasksLeft = result.TasksLeft

	return &ress, nil

}

func (project *ProjectServiceServer) GetProgressofMembers(req *projectpb.GetProgressofMembersReq, stream projectpb.ProjectService_GetProgressofMembersServer) error {

	url := fmt.Sprintf("http://localhost:50005/project/task/stages/count?projectID=%s", req.ProjectID)
	resStages, err := http.Get(url)
	if err != nil {
		helpers.PrintErr(err, "errro happened at calling http method")
		return err
	}

	var list entities.ListofUserProgress
	if err := json.NewDecoder(resStages.Body).Decode(&list); err != nil {
		helpers.PrintErr(err, "errro happened at decoding the json")
		return err
	}

	res, statuses, roleIds, err := project.Usecase.GetProgressofMembers(list, req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "errro happened at GetProgressofMembers usecase")
		return err
	}

	streaam, err := project.UserConn.GetStreamofUserDetails(context.Background())
	if err != nil {
		helpers.PrintErr(err, "errro happened at creating stream")
		return err
	}

	for i := range res.UserAndProgress {

		if err := streaam.Send(&userpb.GetUserDetailsReq{
			UserID: res.UserAndProgress[i].UserID,
			RoleID: uint32(roleIds[i]),
		}); err != nil {
			helpers.PrintErr(err, "errro happened at sending to stream")
			return err
		}

		usrDetails, err := streaam.Recv()
		if err != nil {
			helpers.PrintErr(err, "errro happened at recieving from stream")
			return err
		}

		if err := stream.Send(&projectpb.GetProgressofMembersRes{
			MemberID:     usrDetails.UserID,
			Email:        usrDetails.Email,
			Name:         usrDetails.Name,
			TaskDeadline: res.UserAndProgress[i].TaskDeadline,
			Progress:     res.UserAndProgress[i].Progress,
			TaskStatus:   statuses[i],
			Role:         usrDetails.Role,
		}); err != nil {
			helpers.PrintErr(err, "errro happened at sending to stream")
			return err
		}
	}

	return nil
}

func (project *ProjectServiceServer) GetProjectProgress(ctx context.Context, req *projectpb.GetProjectProgressReq) (*projectpb.GetProjectProgressRes, error) {

	url := fmt.Sprintf("http://localhost:50005/project/task/stages/count?projectID=%s", req.ProjectID)
	resStages, err := http.Get(url)
	if err != nil {
		helpers.PrintErr(err, "errro happened at calling http method")
		return nil, err
	}

	var list entities.ListofUserProgress
	if err := json.NewDecoder(resStages.Body).Decode(&list); err != nil {
		helpers.PrintErr(err, "errro happened at decoding the json")
		return nil, err
	}

	res, err := project.Usecase.GetProjectProgress(req.ProjectID, list)
	if err != nil {
		helpers.PrintErr(err, "errro happened at GetProjectProgress usecase")
		return nil, err
	}

	return &projectpb.GetProjectProgressRes{
		ProjectID:            res.ProjectID,
		Progress:             res.Progress,
		Deadline:             res.Deadline,
		LiveMembers:          res.LiveMembers,
		TaskCompletedMembers: res.TaskCompletedMembers,
		TaskCriticalMembers:  res.TaskCriticalMembers,
	}, nil
}

func (project *ProjectServiceServer) MarkProgressofNonTechnical(ctx context.Context, req *projectpb.MarkProgressofNonTechnicalReq) (*emptypb.Empty, error) {

	if err := project.Usecase.InsertNonTechnicalTasks(entities.NonTechnicalTaskDetials{
		UserID:      req.MemberID,
		ProjectID:   req.ProjectID,
		Task:        req.CompletedTask,
		Description: req.Description,
	}); err != nil {
		helpers.PrintErr(err, "errro happened at InsertNonTechnicalTasks usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) IsMemberAccepted(ctx context.Context, req *projectpb.IsMemberAcceptedReq) (*emptypb.Empty, error) {

	if err := project.Usecase.IsMemberAccepted(req.UserID, req.ProjectID); err != nil {
		helpers.PrintErr(err, "errro happened at IsMemberAccepted usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) GetLiveProjects(req *projectpb.GetLiveProjectsReq, stream projectpb.ProjectService_GetLiveProjectsServer) error {

	res, err := project.Usecase.GetLiveProjectsofCompany(req.CompanyID)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetLiveProjectsofCompany usecase")
		return err
	}

	streaam, err := project.CompanyConn.GetStreamofClients(context.Background())
	if err != nil {
		helpers.PrintErr(err, "error happened at GetStreamofClients")
		return err
	}

	fmt.Println(res)

	for _, v := range res {
		if err = streaam.Send(&companypb.GetStreamofClientsReq{
			CompanyID: req.CompanyID,
			ProjectID: v.ProjectID,
		}); err != nil {
			helpers.PrintErr(err, "error happened at sending to stream")
			return err
		}

		id, err := streaam.Recv()
		if err != nil {
			helpers.PrintErr(err, "error happened at recieving from stream")
			return err
		}

		if err = stream.Send(&projectpb.GetLiveProjectsRes{
			ProjectID:          v.ProjectID,
			ProjectUsername:    v.ProjectUsername,
			ProjectDescription: v.ProjectDescription,
			MembersWorking:     uint32(v.Members),
			ClientID:           id.ClientID,
		}); err != nil {
			helpers.PrintErr(err, "error happened at sendin to stream")
			return err
		}
	}

	return nil
}

func (project *ProjectServiceServer) GetStreamofProjectDetails(stream projectpb.ProjectService_GetStreamofProjectDetailsServer) error {

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			helpers.PrintErr(err, "error happened at sending to stream")
			return err
		}
		res, err := project.Usecase.GetProjectDetails(req.ProjectID, "")
		if err != nil {
			helpers.PrintErr(err, "error happened at GetProjectDetails usecase")
			return err
		}

		if err := stream.Send(&projectpb.GetProjectDetailesRes{
			ProjectID:       res.ProjectID,
			ProjectUsername: res.ProjectUsername,
			Aim:             res.Aim,
		}); err != nil {
			helpers.PrintErr(err, "error happened at sending to stream")
			return err
		}
	}

	return nil
}

func (project *ProjectServiceServer) GetCompletedMembers(req *projectpb.GetCompletedMembersReq, stream projectpb.ProjectService_GetCompletedMembersServer) error {

	url := fmt.Sprintf("http://localhost:50005/project/task/stages/count?projectID=%s", req.ProjectID)
	resStages, err := http.Get(url)
	if err != nil {
		helpers.PrintErr(err, "errro happened at calling http method")
		return err
	}

	var list entities.ListofUserProgress
	if err := json.NewDecoder(resStages.Body).Decode(&list); err != nil {
		helpers.PrintErr(err, "errro happened at decoding the json")
		return err
	}

	users, err := project.Usecase.GetCompletedMembers(req.ProjectID, list, true)
	streaam, err := project.UserConn.GetStreamofUserDetails(context.TODO())
	if err != nil {
		helpers.PrintErr(err, "erorr happened at GetStreamofUserDetails")
	}

	for _, v := range users {

		if err = streaam.Send(&userpb.GetUserDetailsReq{
			UserID: v.UserID,
		}); err != nil {
			helpers.PrintErr(err, "error happened at sending to stream")
			return err
		}

		details, err := streaam.Recv()
		if err != nil {
			helpers.PrintErr(err, "error happened at recieving from stream")
		}

		if err = stream.Send(&projectpb.GetCompletedMembersRes{
			UserID:     details.UserID,
			Email:      details.Email,
			Name:       details.Name,
			IsVerified: v.IsVerified,
		}); err != nil {
			helpers.PrintErr(err, "error hsppened at sending to stream")
			return err
		}
	}

	return nil
}

func (project *ProjectServiceServer) GetCriticalMembers(req *projectpb.GetCriticalMembersReq, stream projectpb.ProjectService_GetCriticalMembersServer) error {

	url := fmt.Sprintf("http://localhost:50005/project/task/stages/count?projectID=%s", req.ProjectID)
	resStages, err := http.Get(url)
	if err != nil {
		helpers.PrintErr(err, "errro happened at calling http method")
		return err
	}

	var list entities.ListofUserProgress
	if err := json.NewDecoder(resStages.Body).Decode(&list); err != nil {
		helpers.PrintErr(err, "errro happened at decoding the json")
		return err
	}

	users, err := project.Usecase.GetCompletedMembers(req.ProjectID, list, false)
	streaam, err := project.UserConn.GetStreamofUserDetails(context.TODO())
	if err != nil {
		helpers.PrintErr(err, "erorr happened at GetStreamofUserDetails")
	}

	for _, v := range users {

		if err = streaam.Send(&userpb.GetUserDetailsReq{
			UserID: v.UserID,
		}); err != nil {
			helpers.PrintErr(err, "error happened at sending to stream")
			return err
		}

		details, err := streaam.Recv()
		if err != nil {
			helpers.PrintErr(err, "error happened at recieving from stream")
		}

		if err = stream.Send(&projectpb.GetCriticalMembersRes{
			UserID: details.UserID,
			Email:  details.Email,
			Name:   details.Name,
		}); err != nil {
			helpers.PrintErr(err, "error hsppened at sending to stream")
			return err
		}
	}

	return nil
}

func (project *ProjectServiceServer) RaiseIssue(ctx context.Context, req *projectpb.RaiseIssueReq) (*emptypb.Empty, error) {

	if err := project.Usecase.RaiseIssue(entities.Issues{
		ProjectID:   req.ProjectID,
		UserID:      req.MemberID,
		Description: req.Description,
		RaiserID:    req.RaiserID,
	}); err != nil {
		helpers.PrintErr(err, "error happened at RaiseIssue usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) GetIssues(ctx context.Context, req *projectpb.GetIssuesReq) (*projectpb.GetIssuesRes, error) {

	res, err := project.Usecase.GetIssuesofMember(req.ProjectID, req.MemberID)
	if err != nil {
		helpers.PrintErr(err, "erorr happened at GetIssuesofMember usecase")
		return nil, err
	}

	return &projectpb.GetIssuesRes{
		RaisedBy:    res.RaiserID,
		Description: res.Description,
	}, nil
}

func (proj *ProjectServiceServer) GetIssuesofProject(req *projectpb.GetIssuesofProjectReq, stream projectpb.ProjectService_GetIssuesofProjectServer) error {

	res, err := proj.Usecase.GetIssuesofProject(req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetIssuesofProject usecase")
		return err
	}

	streaam, err := proj.UserConn.GetStreamofUserDetails(context.TODO())
	if err != nil {
		helpers.PrintErr(err, "error happened at GetStreamofUserDetails usecase")
		return err
	}

	for _, v := range res {

		if err = streaam.Send(&userpb.GetUserDetailsReq{
			UserID: v.UserID,
		}); err != nil {
			helpers.PrintErr(err, "error happened at sending to stream")
			return err
		}

		details, err := streaam.Recv()
		if err != nil {
			helpers.PrintErr(err, "error happened at recieving from sttream")
			return err
		}

		if err = stream.Send(&projectpb.GetIssuesofProjectRes{
			MemberID:    details.UserID,
			Email:       details.Email,
			Name:        details.Name,
			RaisedBy:    v.RaiserID,
			Description: v.Description,
		}); err != nil {
			helpers.PrintErr(err, "error happened at sending to stream")
			return err
		}
	}

	return nil
}

func (proj *ProjectServiceServer) RateTask(ctx context.Context, req *projectpb.RateTaskReq) (*emptypb.Empty, error) {

	if err := proj.Usecase.RateTask(entities.Ratings{
		ProjectID: req.ProjectID,
		UserID:    req.MemberID,
		Rating:    req.Rating,
		Feedback:  req.Feedback,
	}); err != nil {
		helpers.PrintErr(err, "error happened at RateTask usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (proj *ProjectServiceServer) GetfeedBackforTask(ctx context.Context, req *projectpb.GetfeedBackforTaskReq) (*projectpb.GetfeedBackforTaskRes, error) {

	res, err := proj.Usecase.GetRating(req.ProjectID, req.MemberID)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetRating usecase")
		return nil, err
	}

	return &projectpb.GetfeedBackforTaskRes{
		Rating:   res.Rating,
		Feedback: res.Feedback,
	}, nil
}

func (proj *ProjectServiceServer) RequestforDeadlineExtension(ctx context.Context, req *projectpb.RequestforDeadlineExtensionReq) (*emptypb.Empty, error) {

	time, err := time.Parse("2006-01-02", req.ExetendTo)
	if err != nil {
		helpers.PrintErr(err, "error happened at parsing the time")
		return &emptypb.Empty{}, err
	}

	if err = proj.Usecase.AskExtension(entities.Extensions{
		ProjectID:   req.ProjectID,
		UserID:      req.MemberID,
		ExtendTo:    time,
		Description: req.Description,
	}); err != nil {
		helpers.PrintErr(err, "error happened at AskExtension usecase")
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (proj *ProjectServiceServer) GetExtensionRequests(req *projectpb.GetExtensionRequestsReq, stream projectpb.ProjectService_GetExtensionRequestsServer) error {

	res, err := proj.Usecase.GetExtensionRequestsinaProject(req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "error happened at GetExtensionRequestsinaProject usecase")
		return err
	}

	for _, v := range res {

		if err := stream.Send(&projectpb.GetExtensionRequestsRes{
			ID:          uint32(v.ID),
			MemberID:    v.UserID,
			ExetendTo:   v.ExtendTo.String(),
			Description: v.Description,
		}); err != nil {
			helpers.PrintErr(err, "error happened at sending to stream")
			return err
		}
	}

	return nil
}

func (proj *ProjectServiceServer) GrantExtension(ctx context.Context, req *projectpb.GrantExtensionReq) (*emptypb.Empty, error) {

	if err := proj.Usecase.ApproveExtensionRequest(uint(req.ID), req.IsApproved); err != nil {
		helpers.PrintErr(err, "error happened at ApproveExtensionRequest usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (proj *ProjectServiceServer) VerifyTaskCompletion(ctx context.Context, req *projectpb.VerifyTaskCompletionReq) (*emptypb.Empty, error) {

	if err := proj.Usecase.VerifyTaskCompletion(req.ProjectID, req.MemberID, req.Verified); err != nil {
		helpers.PrintErr(err, "error happened at VerifyTaskCompletion usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (proj *ProjectServiceServer) GetVerifiedTasks(req *projectpb.GetVerifiedTasksReq, stream projectpb.ProjectService_GetVerifiedTasksServer) error {

	res, err := proj.Usecase.GetVerifiedTasks(req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "error happenede at GetVerifiedTasks usecase")
		return err
	}

	for _, v := range res {

		if err = stream.Send(&projectpb.GetVerifiedTasksRes{
			MemberID: v.MemberID,
			Rating:   v.Rating,
			Feedback: v.Feedback,
		}); err != nil {
			helpers.PrintErr(err, "error happened at sending to stream")
			return err
		}

	}

	return nil
}

func (project *ProjectServiceServer) DropProject(ctx context.Context, req *projectpb.DropProjectReq) (*emptypb.Empty, error) {

	if err := project.Usecase.DropProject(req.ProjectID); err != nil {
		helpers.PrintErr(err, "error happpneed at DropProject usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) TerminateProjectMembers(ctx context.Context, req *projectpb.TerminateProjectMembersReq) (*emptypb.Empty, error) {

	if err := project.Usecase.TerminateProjectMembers(req.MemberID, req.MemberID); err != nil {
		helpers.PrintErr(err, "error happened at TerminateProjectMembers usecase")
		return nil, err
	}

	if err := project.Cache.Del(ctx, req.ProjectID+" "+req.MemberID).Err(); err != nil {
		helpers.PrintErr(err, "eroror happened at clearing cache")
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) EditProjectDetails(ctx context.Context, req *projectpb.EditProjectDetailsReq) (*emptypb.Empty, error) {

	if err := project.Usecase.EditProject(entities.Credentials{
		ProjectID:       req.ProjectID,
		Description:     req.Description,
		ProjectUsername: req.ProjectUsername,
		Aim:             req.Aim,
		Name:            req.Name,
	}); err != nil {
		helpers.PrintErr(err, "error happened at EditProject usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) EditMember(ctx context.Context, req *projectpb.EditMemberReq) (*emptypb.Empty, error) {

	if err := project.Usecase.EditMember(entities.Members{
		MemberID:     req.MemberID,
		PermissionID: uint(req.PermissionID),
		RoleID:       uint(req.RoleID),
		ProjectID:    req.ProjectID,
	}); err != nil {
		helpers.PrintErr(err, "error happened at EditMember usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) EditFeedback(ctx context.Context, req *projectpb.EditFeedbackReq) (*emptypb.Empty, error) {

	if err := project.Usecase.EditFeedback(entities.Ratings{
		ProjectID: req.ProjectID,
		UserID:    req.MemberID,
		Rating:    req.Rating,
		Feedback:  req.Feedback,
	}); err != nil {
		helpers.PrintErr(err, "error happened at EditFeedback usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) DeleteFeedback(ctx context.Context, req *projectpb.DeleteFeedbackReq) (*emptypb.Empty, error) {

	if err := project.Usecase.DeleteFeedback(req.ProjectID, req.MemberID); err != nil {
		helpers.PrintErr(err, "error happened at DeleteFeedback usecase")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) GetUserStat(ctx context.Context, req *projectpb.GetUserStatReq) (*projectpb.GetUserStatRes, error) {

	res, err := project.Usecase.IsMemberExists(req.UserID, req.ProjectID)
	if err != nil {
		helpers.PrintErr(err, "error happened at IsMemberExists usecase")
		return nil, err
	}

	if !res {

		ress, err := project.Usecase.IsOwner(req.UserID, req.ProjectID)
		if err != nil {
			helpers.PrintErr(err, "error happened at IsOwner usecase")
			return nil, err
		}

		if !ress {
			return &projectpb.GetUserStatRes{
				IsAcceptable: false,
			}, nil
		}
		return &projectpb.GetUserStatRes{
			IsAcceptable: true,
		}, nil

	}

	return &projectpb.GetUserStatRes{
		IsAcceptable: true,
	}, nil
}
