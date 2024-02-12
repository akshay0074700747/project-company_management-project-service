package services

import (
	"log"
	"net"

	"github.com/akshay0074700747/projectandCompany_management_protofiles/pb/projectpb"
	"google.golang.org/grpc"
)

type ProjectEngine struct {
	Srv projectpb.ProjectServiceServer
}

func NewProjectEngine(srv projectpb.ProjectServiceServer) *ProjectEngine {
	return &ProjectEngine{
		Srv: srv,
	}
}
func (engine *ProjectEngine) Start(addr string) {

	server := grpc.NewServer()
	projectpb.RegisterProjectServiceServer(server, engine.Srv)

	listener, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	log.Printf("Project Server is listening...")

	if err = server.Serve(listener); err != nil {
		log.Fatalf("Failed to listen on port %s: %v", addr, err)
	}

}
