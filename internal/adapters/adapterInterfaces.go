package adapters

import "github.com/akshay0074700747/project-company_management-project-service/entities"

type ProjectAdapterInterfaces interface {
	CreateProject(entities.Credentials, string) (entities.Credentials, error)
	insertIntoCompanyBased(entities.Companies) error
	IsProjectUsernameExists(string) (bool, error)
	AddMember(entities.Members) error
	IsMemberExists(id string) (bool, error)
	AcceptInvitation(entities.Members) error
	GetProjectInvites(string) ([]entities.ProjectInviteUsecase, error)
}
