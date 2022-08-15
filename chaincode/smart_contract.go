package chaincode

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"

	"github.com/hyperledger/fabric-chaincode-go/shim"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Transaction.
type SmartContract struct {
	contractapi.Contract
}

// Transaction describes basic details of what makes up a simple transaction.
// Insert struct field in alphabetic order => to achieve determinism across languages.
// golang keeps the order when marshal to json but doesn't order automatically.
type Transaction []byte

// InitLedger adds a base set of transactions to the ledger.
func (s *SmartContract) InitLedger(tci contractapi.TransactionContextInterface) error {
	transactions := []Transaction{
		[]byte("test"),
	}
	for _, tx := range transactions {
		txKey := hex.EncodeToString(tx)
		if err := tci.GetStub().PutState(txKey, tx); err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
	}
	return nil
}

// CreateTX issues a new transaction to the world state with given details.
func (s *SmartContract) CreateTX(ctx contractapi.TransactionContextInterface, hash string) (string, error) {
	tx, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}
	exists, err := s.TXExists(ctx, hash)
	if err != nil {
		return "", err
	}
	if exists {
		return "", fmt.Errorf("the transaction %s already exists", hash)
	}
	return hash, ctx.GetStub().PutState(hash, tx)
}

//// ReadTX returns the transaction stored in the world state with given id.
//func (s *SmartContract) ReadTX(ctx contractapi.TransactionContextInterface, id string) (*Transaction, error) {
//	tx, err := ctx.GetStub().GetState(id)
//	if err != nil {
//		return nil, fmt.Errorf("failed to read from world state: %v", err)
//	}
//	if tx == nil {
//		return nil, fmt.Errorf("the transaction %s does not exist", id)
//	}
//	if err != nil {
//		return nil, err
//	}
//	return tx, nil
//}

// DeleteTX deletes a given transaction from the world state.
func (s *SmartContract) DeleteTX(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.TXExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the transaction %s does not exist", id)
	}
	return ctx.GetStub().DelState(id)
}

// TXExists returns true when transaction with given ID exists in world state.
func (s *SmartContract) TXExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	transactionJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return transactionJSON != nil, nil
}

// GetAllTXs returns all transactions found in world state.
func (s *SmartContract) GetAllTXs(ctx contractapi.TransactionContextInterface) (txList []*Transaction, err error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all transactions in the chaincode namespace.
	var iter shim.StateQueryIteratorInterface
	iter, err = ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return
	}
	defer func(resultsIterator shim.StateQueryIteratorInterface) {
		err = resultsIterator.Close()
	}(iter)

	var response *queryresult.KV
	for iter.HasNext() {
		response, err = iter.Next()
		if err != nil {
			return
		}
		var transaction Transaction
		err = json.Unmarshal(response.Value, &transaction)
		if err != nil {
			return
		}
		txList = append(txList, &transaction)
	}
	return
}
