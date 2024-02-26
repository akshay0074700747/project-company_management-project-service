package entities

import "time"

type ProjectInviteUsecase struct {
	ProjectID       string
	AcceptedMembers uint
}

type ProjectDetailsUsecase struct {
	ProjectID       string
	ProjectUsername string
	Aim             string
	Members         uint
}

type ProjectMembersUsecase struct {
	MemberID   string
	IsAccepted bool
}

type TaskDta struct {
	UserID      string    `json:"user_id"`
	ProjectID   string    `json:"project_id"`
	Task        string    `json:"task"`
	Description string    `json:"description"`
	File        []byte    `json:"file"`
	Stages      int       `json:"stages"`
	Deadline    time.Time `json:"deadline"`
}

type StageRes struct {
	Stages  int             `json:"stages"`
	Details []StagesDetails `json:"details"`
}

type StagesDetails struct {
	Key         string `json:"key" bson:"key"`
	Description string `json:"description" bson:"description"`
	Filename    string `json:"file_name" bson:"filename"`
}

type UserProgressUsecaseRes struct {
	MemberID       string
	ProjectID      string
	TaskDeadline   string
	TasksCompleted uint32
	TasksLeft      uint32
	Progress       string
}

type ListofUserProgress struct {
	UserAndProgress []UserProgress `json:"user_and_progress"`
}

type UserProgress struct {
	UserID       string `json:"user_id"`
	Stages       int    `json:"stages"`
	TaskDeadline string
	Progress     string
	TaskStatus   string
}

type GetProjectProgressUsecase struct {
	ProjectID            string
	Progress             string
	Deadline             string
	LiveMembers          uint32
	TaskCompletedMembers uint32
	TaskCriticalMembers  uint32
}

type GetProjectMembersUsecase struct {
	UserID       string
	RoleID       uint
	PermissionID uint
}
