package helpers

import (
	"fmt"
)

func PrintErr(err error, messge string) {
	fmt.Println(messge, err)
}

func PrintMsg(msg string) {
	fmt.Println(msg)
}
