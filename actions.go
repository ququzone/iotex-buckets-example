package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
)

const (
	host = "api.iotex.one:443"
)

func main() {
	// Create grpc connection
	conn, err := iotex.NewDefaultGRPCConn(host)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// create client
	c := iotex.NewReadOnlyClient(iotexapi.NewAPIServiceClient(conn))

	res, err := c.API().GetActions(context.Background(), &iotexapi.GetActionsRequest{
		Lookup: &iotexapi.GetActionsRequest_ByAddr{
			ByAddr: &iotexapi.GetActionsByAddressRequest{
				Address: "io1rl2z079wqd6aug8a0xl9fcg288v6rzkxyw2m6l",
				Start:   0,
				Count:   1000,
			},
		},
	})
	if err != nil {
		log.Fatalf("get actions error: %v\n", err)
	}
	addresses := make(map[string]bool, 0)
	baseAmount, _ := new(big.Int).SetString("500000000000000000000", 10)
	for _, action := range res.ActionInfo {
		if action.Sender != "io1rl2z079wqd6aug8a0xl9fcg288v6rzkxyw2m6l" &&
			action.GetAction().GetCore().GetTransfer() != nil {
			transfer := action.GetAction().Core.GetTransfer()
			amount, _ := new(big.Int).SetString(transfer.Amount, 10)
			if amount.Cmp(baseAmount) >= 0 {
				if !addresses[action.Sender] {
					addresses[action.Sender] = true
				}
				fmt.Printf("%s,%s,%s\n", action.Sender, amount.String(), action.ActHash)
			}
		}
	}

	fmt.Println("---------------")
	for addr := range addresses {
		fmt.Println(addr)
	}
}
