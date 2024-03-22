package injectdependency

import (
	"github.com/akshay0074700747/project-company_management-project-service/config"
	"github.com/akshay0074700747/project-company_management-project-service/db"
	"github.com/akshay0074700747/project-company_management-project-service/internal/adapters"
	"github.com/akshay0074700747/project-company_management-project-service/internal/services"
	"github.com/akshay0074700747/project-company_management-project-service/internal/usecases"
	"github.com/akshay0074700747/project-company_management-project-service/notify"
)

func Initialize(cfg config.Config) *services.ProjectEngine {

	dbPostgres := db.ConnectDB(cfg)
	minioDB := db.ConnectMinio(cfg)
	adapter := adapters.NewProjectAdapter(dbPostgres, minioDB)
	usecase := usecases.NewProjectUseCases(adapter)
	server := services.NewProjectServiceServer(usecase, "user-service:50001", "company-service:50003", "Emailsender", notify.InitEmailNotifier())
	go server.StartConsuming()

	return services.NewProjectEngine(server)
}
