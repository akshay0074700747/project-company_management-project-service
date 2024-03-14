package adapters

import (
	"context"
	"io"
	"time"

	"github.com/akshay0074700747/project-company_management-project-service/entities"
	"github.com/minio/minio-go/v7"
)

type ProjectAdapterInterfaces interface {
	CreateProject(entities.Credentials, string) (entities.Credentials, error)
	insertIntoCompanyBased(entities.Companies) error
	IsProjectUsernameExists(string) (bool, error)
	AddMember(entities.Members) error
	IsMemberExists(string, string) (bool, error)
	AcceptInvitation(entities.Members) error
	GetProjectInvites(string) ([]entities.ProjectInviteUsecase, error)
	AddOwner(string, string) error
	GetNoofMembers(string) (uint, error)
	GetProjectDetails(string) (entities.Credentials, error)
	GetProjectMembers(string) ([]entities.GetProjectMembersUsecase, error)
	AddMemberStatueses(string) error
	DenyInvitation(entities.Members) error
	InsertTasktoMinio(context.Context, string, io.Reader, int64, minio.PutObjectOptions) error
	InsertTaskDetails(entities.TaskAssignations) error
	GetTaskDetails(string, string) (entities.TaskAssignations, error)
	GetTaskFromMinio(context.Context, string, minio.GetObjectOptions) ([]byte, error)
	InsertStatuses(entities.TaskStatuses) error
	GetRoleID(string) (uint, error)
	LogintoProject(string, string) (entities.Members, error)
	GetIDfromName(string) (string, error)
	GetStagesandDeadline([]string, string) ([]entities.TaskAssignations, error)
	GetTaskStatuses([]int) ([]string, error)
	GetListofRoleIds([]string, string) ([]uint, error)
	InsertNonTechnicalTasks(entities.NonTechnicalTaskDetials) error
	GetCountofLivemembers(string) (int, error)
	GetProjectDeadline(string) (time.Time, error)
	GetStagesofProgress(string, string) (int, error)
	IsOwner(string, string) (bool, error)
	IsCompanyBased(string) (bool, string, error)
	MemberState(string, string) (string, error)
	GetLiveProjectsofCompany(string) ([]entities.GetLiveProjectsUsecase, error)
	GetCompletedMembers(string, []string) ([]entities.GetCompletedMemebersAdapter, error)
	RaiseIssue(entities.Issues)(error)
	GetIssuesofMember(string,string)(entities.Issues,error)
	GetIssuesofProject(string)([]entities.Issues, error)
	RateTask(entities.Ratings)(error)
	GetRating(string,string)(entities.Ratings,error)
	AskExtension(entities.Extensions)(error)
	GetExtensionRequestsinaProject(string)([]entities.Extensions,error)
	ApproveExtensionRequest(uint,bool)(error)
	VerifyTaskCompletion(string,string,bool)(error)
	GetVerifiedTasks(string)([]entities.VerifiedTasksUsecase,error)
	DropProject(string)(error)
	EditProject(entities.Credentials)(error)
	EditMember(entities.Members)(error)
	EditFeedback(entities.Ratings)(error)
	DeleteFeedback(string,string)(error)
	GetCountMembers(string) (uint,error)
}
