package helpers

import (
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func PrintErr(err error, messge string) {
	fmt.Println(messge, err)
}

func PrintMsg(msg string) {
	fmt.Println(msg)
}

func DialGrpc(addr string) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, grpc.WithInsecure())
}

func GenUuid() string {
	return uuid.New().String()
}
