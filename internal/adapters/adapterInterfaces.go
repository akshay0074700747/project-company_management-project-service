package adapters

import "github.com/akshay0074700747/project-company_management-project-service/entities"

type ProjectAdapterInterfaces interface {
	CreateProject(entities.Credentials) (entities.Credentials, error)
}
