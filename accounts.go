package main

import (
	"fmt"
	"log"

	"github.com/iotexproject/iotex-antenna-go/v2/account"
)

func main() {
	acc, err := account.NewAccount()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(acc.Address())
	fmt.Println(acc.PrivateKey().HexString())
}
