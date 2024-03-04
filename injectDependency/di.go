package injectdependency

import (
	"github.com/akshay0074700747/project-company_management-project-service/config"
	"github.com/akshay0074700747/project-company_management-project-service/db"
	"github.com/akshay0074700747/project-company_management-project-service/internal/adapters"
	"github.com/akshay0074700747/project-company_management-project-service/internal/services"
	"github.com/akshay0074700747/project-company_management-project-service/internal/usecases"
)

func Initialize(cfg config.Config) *services.ProjectEngine {

	dbPostgres := db.ConnectDB(cfg)
	minioDB := db.ConnectMinio(cfg)
	adapter := adapters.NewProjectAdapter(dbPostgres, minioDB)
	usecase := usecases.NewProjectUseCases(adapter)
	server := services.NewProjectServiceServer(usecase, ":50001", ":50003")
	go server.StartConsuming()

	return services.NewProjectEngine(server)
}
