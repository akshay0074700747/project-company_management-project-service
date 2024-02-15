package usecases

import "github.com/akshay0074700747/project-company_management-project-service/entities"

type ProjectUsecaseInterfaces interface {
	CreateProject(entities.Credentials, string) (entities.Credentials, error)
	Addmembers(entities.Members) error
	AcceptInvitation(entities.Members) error
	GetProjectInvites(string) ([]entities.ProjectInviteUsecase, error)
}
