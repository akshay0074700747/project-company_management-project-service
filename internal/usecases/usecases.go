package usecases

import "github.com/akshay0074700747/project-company_management-project-service/internal/adapters"


type ProjectUseCases struct {
	Adapter adapters.ProjectAdapterInterfaces
}

func NewProjectUseCases(adapter adapters.ProjectAdapterInterfaces) *ProjectUseCases {
	return &ProjectUseCases{
		Adapter: adapter,
	}
}
