package test

import (
	"testing"

	"github.com/akshay0074700747/project-company_management-project-service/entities"
	mock_adapters "github.com/akshay0074700747/project-company_management-project-service/internal/adapters/mockAdapters"
	"github.com/akshay0074700747/project-company_management-project-service/internal/usecases"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateProject(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	adapter := mock_adapters.NewMockProjectAdapterInterfaces(ctrl)

	tests := []struct {
		name                        string
		mockIsProjectUsernameExists func(string) (bool, error)
		mockCreateProject           func(entities.Credentials, string) (entities.Credentials, error)
		mockAddOwner                func(string, string) error
		projectRequest              struct {
			credentials entities.Credentials
			companyID   string
			ownerID     string
		}
		wantError      bool
		expectedResult entities.Credentials
	}{
		{
			name: "Success",
			mockIsProjectUsernameExists: func(s string) (bool, error) {
				return "itsunique" == s, nil
			},
			mockCreateProject: func(c entities.Credentials, s string) (entities.Credentials, error) {
				return c, nil
			},
			mockAddOwner: func(s1, s2 string) error {
				return nil
			},
			projectRequest: struct {
				credentials entities.Credentials
				companyID   string
				ownerID     string
			}{
				credentials: entities.Credentials{
					ProjectUsername: "hsgafgrjfkejfh",
					Name:            "hdbckhdj",
					Aim:             "bvjhqfvjklshvk",
					Description:     "cjhsdacshaljc",
					IsCompanybased:  true,
				},
				companyID: "bsckywrqt46yb8e37q4gffdydbq347",
				ownerID:   "6438gfh73dhud932892dwqd",
			},
			wantError: false,
			expectedResult: entities.Credentials{
				ProjectUsername: "hsgafgrjfkejfh",
				Name:            "hdbckhdj",
				Aim:             "bvjhqfvjklshvk",
				Description:     "cjhsdacshaljc",
				IsCompanybased:  true,
			},
		},
		{
			name: "Fail",
			mockIsProjectUsernameExists: func(s string) (bool, error) {
				return "itsunique" == s, nil
			},
			mockCreateProject: func(c entities.Credentials, s string) (entities.Credentials, error) {
				return c, nil
			},
			mockAddOwner: func(s1, s2 string) error {
				return nil
			},
			projectRequest: struct {
				credentials entities.Credentials
				companyID   string
				ownerID     string
			}{
				credentials: entities.Credentials{
					ProjectUsername: "",
					Name:            "hdbckhdj",
					Aim:             "bvjhqfvjklshvk",
					Description:     "cjhsdacshaljc",
					IsCompanybased:  true,
				},
				companyID: "bsckywrqt46yb8e37q4gffdydbq347",
				ownerID:   "6438gfh73dhud932892dwqd",
			},
			wantError: true,
			expectedResult: entities.Credentials{
				ProjectUsername: "hsgafgrjfkejfh",
				Name:            "hdbckhdj",
				Aim:             "bvjhqfvjklshvk",
				Description:     "cjhsdacshaljc",
				IsCompanybased:  true,
			},
		},
		{
			name: "Fail",
			mockIsProjectUsernameExists: func(s string) (bool, error) {
				return "itsunique" == s, nil
			},
			mockCreateProject: func(c entities.Credentials, s string) (entities.Credentials, error) {
				return c, nil
			},
			mockAddOwner: func(s1, s2 string) error {
				return nil
			},
			projectRequest: struct {
				credentials entities.Credentials
				companyID   string
				ownerID     string
			}{
				credentials: entities.Credentials{
					ProjectUsername: "hsgafgrjfkejfh",
					Name:            "",
					Aim:             "bvjhqfvjklshvk",
					Description:     "cjhsdacshaljc",
					IsCompanybased:  true,
				},
				companyID: "bsckywrqt46yb8e37q4gffdydbq347",
				ownerID:   "6438gfh73dhud932892dwqd",
			},
			wantError: true,
			expectedResult: entities.Credentials{
				ProjectUsername: "hsgafgrjfkejfh",
				Name:            "hdbckhdj",
				Aim:             "bvjhqfvjklshvk",
				Description:     "cjhsdacshaljc",
				IsCompanybased:  true,
			},
		},
	}

	for _, test := range tests {

		if !test.wantError {
			adapter.EXPECT().IsProjectUsernameExists(gomock.Any()).DoAndReturn(test.mockIsProjectUsernameExists).AnyTimes().Times(1)
			adapter.EXPECT().CreateProject(gomock.Any(), gomock.Any()).DoAndReturn(test.mockCreateProject).AnyTimes().Times(1)
			adapter.EXPECT().AddOwner(gomock.Any(), gomock.Any()).DoAndReturn(test.mockAddOwner).AnyTimes().Times(1)
		}

		regUsecase := usecases.NewProjectUseCases(adapter)

		res, err := regUsecase.CreateProject(test.projectRequest.credentials, test.projectRequest.companyID, test.projectRequest.ownerID)
		if test.wantError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, res)
			res.ProjectID = ""
			assert.Equal(t, test.expectedResult, res)
		}
	}
}

func TestAddmembers(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	adapter := mock_adapters.NewMockProjectAdapterInterfaces(ctrl)

	tests := []struct {
		name               string
		mockIsMemberExists func(string, string) (bool, error)
		mockAddMember      func(entities.Members) error
		memberRequest      entities.Members
		wantError          bool
	}{
		{
			name: "Success",
			mockIsMemberExists: func(s1, s2 string) (bool, error) {
				return false, nil
			},
			mockAddMember: func(m entities.Members) error {
				return nil
			},
			memberRequest: entities.Members{
				MemberID:     "gcd6348ygdeg3673e2s",
				ProjectID:    "gf7634b8o2d7b38h2swiojqi93",
				RoleID:       2,
				PermissionID: 3,
			},
			wantError: false,
		},
		{
			name: "Fail",
			mockIsMemberExists: func(s1, s2 string) (bool, error) {
				return true, nil
			},
			mockAddMember: func(m entities.Members) error {
				return nil
			},
			memberRequest: entities.Members{
				MemberID:     "gcd6348ygdeg3673e2s",
				ProjectID:    "gf7634b8o2d7b38h2swiojqi93",
				RoleID:       2,
				PermissionID: 3,
			},
			wantError: true,
		},
		{
			name: "Success",
			mockIsMemberExists: func(s1, s2 string) (bool, error) {
				return false, nil
			},
			mockAddMember: func(m entities.Members) error {
				return nil
			},
			memberRequest: entities.Members{
				MemberID:     "gcd6348ygddsc3eg3673e2s",
				ProjectID:    "gf7634b8o2wdwd7b38h2swiojqi93",
				RoleID:       5,
				PermissionID: 2,
			},
			wantError: false,
		},
	}

	for _, test := range tests {

		if !test.wantError {
			adapter.EXPECT().AddMember(gomock.Any()).DoAndReturn(test.mockAddMember).AnyTimes().Times(1)
			adapter.EXPECT().IsMemberExists(gomock.Any(), gomock.Any()).DoAndReturn(test.mockIsMemberExists).AnyTimes().Times(1)
		} else {
			adapter.EXPECT().IsMemberExists(gomock.Any(), gomock.Any()).DoAndReturn(test.mockIsMemberExists).AnyTimes().Times(1)
		}

		regUsecase := usecases.NewProjectUseCases(adapter)

		err := regUsecase.Addmembers(test.memberRequest)
		if test.wantError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestAcceptInvitation(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	adapter := mock_adapters.NewMockProjectAdapterInterfaces(ctrl)

	tests := []struct {
		name                 string
		mockAcceptInvitation func(entities.Members) error
		mockDenyInvitation   func(entities.Members) error
		memberRequest        entities.Members
		acceptedRequest      bool
		wantError            bool
	}{
		{
			name: "Success",
			mockAcceptInvitation: func(m entities.Members) error {
				return nil
			},
			mockDenyInvitation: func(m entities.Members) error {
				return nil
			},
			memberRequest: entities.Members{
				MemberID:  "gcd6348ygdeg3673e2s",
				ProjectID: "gf7634b8o2d7b38h2swiojqi93",
			},
			acceptedRequest: true,
			wantError:       false,
		},
		{
			name: "Fail",
			mockAcceptInvitation: func(m entities.Members) error {
				return nil
			},
			mockDenyInvitation: func(m entities.Members) error {
				return nil
			},
			memberRequest: entities.Members{
				MemberID:  "",
				ProjectID: "gf7634b8o2d7b38h2swiojqi93",
			},
			acceptedRequest: true,
			wantError:       true,
		},
		{
			name: "Success",
			mockAcceptInvitation: func(m entities.Members) error {
				return nil
			},
			mockDenyInvitation: func(m entities.Members) error {
				return nil
			},
			memberRequest: entities.Members{
				MemberID:  "gcd6348ygdeg3673e2s",
				ProjectID: "gf7634b8o2d7b38h2swiojqi93",
			},
			acceptedRequest: false,
			wantError:       false,
		},
	}

	for _, test := range tests {

		if !test.wantError && test.acceptedRequest {
			adapter.EXPECT().AcceptInvitation(gomock.Any()).DoAndReturn(test.mockAcceptInvitation).AnyTimes().Times(1)
		} else if !test.wantError && !test.acceptedRequest {
			adapter.EXPECT().DenyInvitation(gomock.Any()).DoAndReturn(test.mockDenyInvitation).AnyTimes().Times(1)
		}

		regUsecase := usecases.NewProjectUseCases(adapter)

		err := regUsecase.AcceptInvitation(test.memberRequest, test.acceptedRequest)
		if test.wantError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
