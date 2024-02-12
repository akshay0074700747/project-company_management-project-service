package services

import (
	"context"

	"github.com/akshay0074700747/project-company_management-project-service/internal/usecases"
	"github.com/akshay0074700747/projectandCompany_management_protofiles/pb/projectpb"
)

type ProjectServiceServer struct {
	Usecase usecases.ProjectUsecaseInterfaces
	projectpb.UnimplementedProjectServiceServer
}

func NewProjectServiceServer(usecase usecases.ProjectUsecaseInterfaces) *ProjectServiceServer {
	return &ProjectServiceServer{
		Usecase: usecase,
	}
}

func (auth *ProjectServiceServer) CreateProject(context.Context, *projectpb.CreateProjectReq) (*projectpb.CreateProjectRes, error) {

}
