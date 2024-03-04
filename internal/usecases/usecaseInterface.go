package usecases

import "github.com/akshay0074700747/project-company_management-project-service/entities"

type ProjectUsecaseInterfaces interface {
	CreateProject(entities.Credentials, string, string) (entities.Credentials, error)
	Addmembers(entities.Members) error
	AcceptInvitation(entities.Members, bool) error
	GetProjectInvites(string) ([]entities.ProjectInviteUsecase, error)
	GetProjectDetails(string, string) (entities.ProjectDetailsUsecase, error)
	GetProjectMembers(string) ([]entities.GetProjectMembersUsecase, error)
	AddMemberStatueses(string) error
	AssignTasks(entities.TaskDta) error
	GetTaskDetails(string, string) (entities.TaskAssignations, error)
	DownloadTask(string) ([]byte, error)
	InsertStatuses(entities.TaskStatuses) error
	GetProgressofMember(entities.UserProgressUsecaseRes) (entities.UserProgressUsecaseRes, error)
	GetRoleID(string) (uint, error)
	LogintoProject(string, string) (entities.Members, error)
	GetProgressofMembers(entities.ListofUserProgress, string) (entities.ListofUserProgress, []string, []uint, error)
	InsertNonTechnicalTasks(entities.NonTechnicalTaskDetials) error
	GetProjectProgress(string,entities.ListofUserProgress)(entities.GetProjectProgressUsecase,error)
	IsOwner(string, string) (bool, error)
	IsCompanyBased(string)(bool,string,error)
	IsMemberAccepted(string,string)(error)
	GetLiveProjectsofCompany(string) ([]entities.GetLiveProjectsUsecase, error)
}
