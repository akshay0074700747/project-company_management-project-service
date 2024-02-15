package services

import (
	"context"
	"errors"

	"github.com/akshay0074700747/project-company_management-project-service/entities"
	"github.com/akshay0074700747/project-company_management-project-service/helpers"
	"github.com/akshay0074700747/project-company_management-project-service/internal/usecases"
	"github.com/akshay0074700747/projectandCompany_management_protofiles/pb/projectpb"
	"github.com/akshay0074700747/projectandCompany_management_protofiles/pb/userpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProjectServiceServer struct {
	Usecase  usecases.ProjectUsecaseInterfaces
	UserConn userpb.UserServiceClient
	projectpb.UnimplementedProjectServiceServer
}

func NewProjectServiceServer(usecase usecases.ProjectUsecaseInterfaces, addr string) *ProjectServiceServer {
	userRes, _ := helpers.DialGrpc(addr)
	return &ProjectServiceServer{
		Usecase:  usecase,
		UserConn: userpb.NewUserServiceClient(userRes),
	}
}

func (project *ProjectServiceServer) CreateProject(tx context.Context, req *projectpb.CreateProjectReq) (*projectpb.CreateProjectRes, error) {

	res, err := project.Usecase.CreateProject(entities.Credentials{
		ProjectUsername: req.ProjectUsername,
		Name:            req.Name,
		Aim:             req.Aim,
		Description:     req.Description,
		IsCompanybased:  req.IsCompanybased,
	}, req.ComapanyId)

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
	}); err != nil {
		helpers.PrintErr(err, "error occured at AcceptInvitation usecase...")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (project *ProjectServiceServer) AddMembers(ctx context.Context, req *projectpb.AddMemberReq) (*emptypb.Empty, error) {

	res, err := project.UserConn.GetByEmail(ctx, &userpb.GetByEmailReq{
		Email: req.Email,
	})

	if err != nil {
		return nil, errors.New("cannot connect to user-service now please try again later")
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
