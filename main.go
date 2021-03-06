package main

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/iotexproject/iotex-antenna-go/v2/account"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"google.golang.org/protobuf/proto"
)

const (
	host = "api.testnet.iotex.one:443"
)

func main() {
	// Create grpc connection
	conn, err := iotex.NewDefaultGRPCConn(host)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Add account by private key
	acc, err := account.HexStringToAccount(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	// create client
	c := iotex.NewAuthedClient(iotexapi.NewAPIServiceClient(conn), acc)

	candidates, err := fetchCandidates(c)
	if err != nil {
		log.Fatal(err)
	}

	bucketId, err := createBucket(c, candidates[0])
	if err != nil {
		log.Fatalf("create bucket error: %v\n", err)
	}
	log.Printf("create bucket #%d\n", bucketId)

	err = addDeposit(c, bucketId)
	if err != nil {
		log.Fatalf("add deposit error: %v\n", err)
	}
	log.Println("add deposit successful")

	err = changeCandidate(c, bucketId, candidates[1])
	if err != nil {
		log.Fatalf("change candidate error: %v\n", err)
	}
	log.Println("change candidate successful")

	err = restakeBucket(c, bucketId)
	if err != nil {
		log.Fatalf("restake bucket error: %v\n", err)
	}
	log.Println("restake bucket  successful")

	// err = unstakeBucket(c, bucketId)
	// if err != nil {
	// 	log.Fatalf("unstake bucket error: %v\n", err)
	// }
	// log.Println("unstake bucket successful")

	bucketIds, err := fetchVoterBuckets(c, c.Account().Address().String())
	if err != nil {
		log.Fatalf("fetch voter buckets error: %v\n", err)
	}
	log.Printf("voter %s has %d buckets\n", c.Account().Address().String(), len(bucketIds))

	bucket, err := fetchBucket(c, bucketIds[0])
	if err != nil {
		log.Fatalf("fetch bucket error: %v\n", err)
	}
	log.Printf("bucket #%d staking %s IOTX\n", bucket.Index, bucket.StakedAmount)
}

func createBucket(c iotex.AuthedClient, candidateName string) (uint64, error) {
	amount, _ := big.NewInt(0).SetString("200000000000000000000", 10)
	hash, err := c.Staking().Create(candidateName, amount, 7, true).
		SetGasLimit(100000).SetGasPrice(big.NewInt(10000000000000)).
		Call(context.Background())
	if err != nil {
		return 0, err
	}

	time.Sleep(10 * time.Second)
	receipt, err := c.API().GetReceiptByAction(context.Background(), &iotexapi.GetReceiptByActionRequest{
		ActionHash: hex.EncodeToString(hash[:]),
	})
	if err != nil {
		return 0, err
	}
	if receipt.ReceiptInfo.Receipt.Status != 1 {
		return 0, errors.New("action revert")
	}

	return new(big.Int).SetBytes(receipt.ReceiptInfo.Receipt.Logs[0].Topics[1]).Uint64(), nil
}

func addDeposit(c iotex.AuthedClient, bucketId uint64) error {
	amount, _ := big.NewInt(0).SetString("10000000000000000000", 10)
	hash, err := c.Staking().AddDeposit(bucketId, amount).
		SetGasLimit(100000).SetGasPrice(big.NewInt(10000000000000)).
		Call(context.Background())
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)
	receipt, err := c.API().GetReceiptByAction(context.Background(), &iotexapi.GetReceiptByActionRequest{
		ActionHash: hex.EncodeToString(hash[:]),
	})
	if err != nil {
		return err
	}
	if receipt.ReceiptInfo.Receipt.Status != 1 {
		return errors.New("action revert")
	}
	return nil
}

func changeCandidate(c iotex.AuthedClient, bucketId uint64, candidateName string) error {
	hash, err := c.Staking().ChangeCandidate(candidateName, bucketId).
		SetGasLimit(100000).SetGasPrice(big.NewInt(10000000000000)).
		Call(context.Background())
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)
	receipt, err := c.API().GetReceiptByAction(context.Background(), &iotexapi.GetReceiptByActionRequest{
		ActionHash: hex.EncodeToString(hash[:]),
	})
	if err != nil {
		return err
	}
	if receipt.ReceiptInfo.Receipt.Status != 1 {
		return errors.New("action revert")
	}
	return nil
}

func restakeBucket(c iotex.AuthedClient, bucketId uint64) error {
	hash, err := c.Staking().Restake(bucketId, 21, false).
		SetGasLimit(100000).SetGasPrice(big.NewInt(10000000000000)).
		Call(context.Background())
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)
	receipt, err := c.API().GetReceiptByAction(context.Background(), &iotexapi.GetReceiptByActionRequest{
		ActionHash: hex.EncodeToString(hash[:]),
	})
	if err != nil {
		return err
	}
	if receipt.ReceiptInfo.Receipt.Status != 1 {
		return errors.New("action revert")
	}
	return nil
}

func unstakeBucket(c iotex.AuthedClient, bucketId uint64) error {
	hash, err := c.Staking().Unstake(bucketId).
		SetGasLimit(100000).SetGasPrice(big.NewInt(10000000000000)).
		Call(context.Background())
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second)
	receipt, err := c.API().GetReceiptByAction(context.Background(), &iotexapi.GetReceiptByActionRequest{
		ActionHash: hex.EncodeToString(hash[:]),
	})
	if err != nil {
		return err
	}
	if receipt.ReceiptInfo.Receipt.Status != 1 {
		return errors.New("action revert")
	}
	return nil
}

func fetchCandidates(c iotex.AuthedClient) ([]string, error) {
	method := &iotexapi.ReadStakingDataMethod{
		Method: iotexapi.ReadStakingDataMethod_CANDIDATES,
	}
	methodBytes, err := proto.Marshal(method)
	if err != nil {
		return nil, err
	}
	argumentsBytes, err := buildReadCandidatesData(0, 100)
	if err != nil {
		return nil, err
	}
	response, err := c.API().ReadState(context.Background(), &iotexapi.ReadStateRequest{
		ProtocolID: []byte("staking"),
		MethodName: methodBytes,
		Arguments:  [][]byte{argumentsBytes},
		Height:     "",
	})
	var result iotextypes.CandidateListV2
	err = proto.Unmarshal(response.Data, &result)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(result.Candidates))
	for i := 0; i < len(names); i++ {
		names[i] = result.Candidates[i].Name
	}

	return names, nil
}

func fetchVoterBuckets(c iotex.AuthedClient, voter string) ([]uint64, error) {
	method := &iotexapi.ReadStakingDataMethod{
		Method: iotexapi.ReadStakingDataMethod_BUCKETS_BY_VOTER,
	}
	methodBytes, err := proto.Marshal(method)
	if err != nil {
		return nil, err
	}
	argumentsBytes, err := buildReadBucketsData(voter, 0, 100)
	if err != nil {
		return nil, err
	}
	response, err := c.API().ReadState(context.Background(), &iotexapi.ReadStateRequest{
		ProtocolID: []byte("staking"),
		MethodName: methodBytes,
		Arguments:  [][]byte{argumentsBytes},
		Height:     "",
	})
	var result iotextypes.VoteBucketList
	err = proto.Unmarshal(response.Data, &result)
	if err != nil {
		return nil, err
	}
	bucketIds := make([]uint64, len(result.Buckets))
	for i := 0; i < len(bucketIds); i++ {
		bucketIds[i] = result.Buckets[i].Index
	}

	return bucketIds, nil
}

func fetchBucket(c iotex.AuthedClient, bucketId uint64) (*iotextypes.VoteBucket, error) {
	method := &iotexapi.ReadStakingDataMethod{
		Method: iotexapi.ReadStakingDataMethod_BUCKETS_BY_INDEXES,
	}
	methodBytes, err := proto.Marshal(method)
	if err != nil {
		return nil, err
	}
	argumentsBytes, err := buildReadBucketData([]uint64{bucketId})
	if err != nil {
		return nil, err
	}
	response, err := c.API().ReadState(context.Background(), &iotexapi.ReadStateRequest{
		ProtocolID: []byte("staking"),
		MethodName: methodBytes,
		Arguments:  [][]byte{argumentsBytes},
		Height:     "",
	})
	var result iotextypes.VoteBucketList
	err = proto.Unmarshal(response.Data, &result)
	if err != nil {
		return nil, err
	}
	if len(result.Buckets) > 0 {
		return result.Buckets[0], nil
	}

	return nil, nil
}

func buildReadCandidatesData(offset, limit uint32) ([]byte, error) {
	arguments := &iotexapi.ReadStakingDataRequest{
		Request: &iotexapi.ReadStakingDataRequest_Candidates_{
			Candidates: &iotexapi.ReadStakingDataRequest_Candidates{
				Pagination: &iotexapi.PaginationParam{
					Offset: offset,
					Limit:  limit,
				},
			},
		},
	}
	return proto.Marshal(arguments)
}

func buildReadBucketData(bucketIds []uint64) ([]byte, error) {
	arguments := &iotexapi.ReadStakingDataRequest{
		Request: &iotexapi.ReadStakingDataRequest_BucketsByIndexes{
			BucketsByIndexes: &iotexapi.ReadStakingDataRequest_VoteBucketsByIndexes{
				Index: bucketIds,
			},
		},
	}
	return proto.Marshal(arguments)
}

func buildReadBucketsData(voter string, offset, limit uint32) ([]byte, error) {
	arguments := &iotexapi.ReadStakingDataRequest{
		Request: &iotexapi.ReadStakingDataRequest_BucketsByVoter{
			BucketsByVoter: &iotexapi.ReadStakingDataRequest_VoteBucketsByVoter{
				VoterAddress: voter,
				Pagination: &iotexapi.PaginationParam{
					Offset: offset,
					Limit:  limit,
				},
			},
		},
	}
	return proto.Marshal(arguments)
}
