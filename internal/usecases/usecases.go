package usecases

import (
	"errors"

	"github.com/akshay0074700747/project-company_management-project-service/entities"
	"github.com/akshay0074700747/project-company_management-project-service/helpers"
	"github.com/akshay0074700747/project-company_management-project-service/internal/adapters"
)

type ProjectUseCases struct {
	Adapter adapters.ProjectAdapterInterfaces
}

func NewProjectUseCases(adapter adapters.ProjectAdapterInterfaces) *ProjectUseCases {
	return &ProjectUseCases{
		Adapter: adapter,
	}
}

func (project *ProjectUseCases) CreateProject(req entities.Credentials, compID string) (entities.Credentials, error) {

	if req.ProjectUsername == "" {
		return entities.Credentials{}, errors.New("the project username cannot be empty")
	}

	if req.Name == "" {
		return entities.Credentials{}, errors.New("the project rname cannot be empty")
	}

	isExisting, err := project.Adapter.IsProjectUsernameExists(req.ProjectUsername)
	if err != nil {
		helpers.PrintErr(err, "error occured at IsProjectUsernameExists adapter")
	}

	if isExisting {
		return entities.Credentials{}, errors.New("the username is already taken")
	}

	res, err := project.Adapter.CreateProject(req, compID)
	if err != nil {
		helpers.PrintErr(err, "error occured at CreateProject adapter")
		return entities.Credentials{}, err
	}

	return res, nil

}

func (project *ProjectUseCases) Addmembers(req entities.Members) error {

	isExists, err := project.Adapter.IsMemberExists(req.MemberID)
	if err != nil {
		helpers.PrintErr(err, "error occures at IsMemberExists adapter")
		return errors.New("there has been an error , please try again later")
	}

	if isExists {
		return errors.New("the member already exists...")
	}

	if err = project.Adapter.AddMember(req); err != nil {
		helpers.PrintErr(err, "error occures at AddMember adapter")
		return errors.New("there has been an error , please try again later")
	}

	return nil
}

func (project *ProjectUseCases) AcceptInvitation(req entities.Members) error {

	if err := project.Adapter.AcceptInvitation(req); err != nil {
		helpers.PrintErr(err, "error occured at AcceptInvitation adapter")
		return err
	}

	return nil
}

func (project *ProjectUseCases) GetProjectInvites(memID string) ([]entities.ProjectInviteUsecase, error) {

	res, err := project.Adapter.GetProjectInvites(memID)
	if err != nil {
		helpers.PrintErr(err, "error at GetProjectInvites adapter")
		return nil, err
	}

	return res, nil
}
