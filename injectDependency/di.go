package injectdependency

import (
	"github.com/akshay0074700747/project-company_management-project-service/config"
	"github.com/akshay0074700747/project-company_management-project-service/db"
	"github.com/akshay0074700747/project-company_management-project-service/internal/adapters"
	"github.com/akshay0074700747/project-company_management-project-service/internal/services"
	"github.com/akshay0074700747/project-company_management-project-service/internal/usecases"
)

func Initialize(cfg config.Config) *services.ProjectEngine {

	db := db.ConnectDB(cfg)
	adapter := adapters.NewProjectAdapter(db)
	usecase := usecases.NewProjectUseCases(adapter)
	server := services.NewProjectServiceServer(usecase)

	return services.NewProjectEngine(server)
}
