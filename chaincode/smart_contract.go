package chaincode

import (
	"crypto/sha256"
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
// Binding: the binding of personal information hash and certificate information hash
// i.e. (person_info_hash || cert_into_hash)
// Timestamp: the timestamp of the transaction
type Transaction struct {
	Binding   string `json:"Binding"`
	Timestamp int64  `json:"Timestamp"`
}

// InitLedger adds a base set of transactions to the ledger.
func (s *SmartContract) InitLedger(tci contractapi.TransactionContextInterface) error {
	transactions := []Transaction{
		{
			Binding:   "test_binding",
			Timestamp: 0,
		},
	}
	sha256Func := sha256.New()
	for _, tx := range transactions {
		// compose transaction key: SHA256 hash of the transaction in JSON format
		txJSON, err := json.Marshal(tx)
		if err != nil {
			return err
		}
		sha256Func.Write(txJSON)
		txKey := hex.EncodeToString(sha256Func.Sum(nil))
		err = tci.GetStub().PutState(txKey, txJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
		sha256Func.Reset()
	}
	return nil
}

// CreateTX issues a new transaction to the world state with given details.
func (s *SmartContract) CreateTX(ctx contractapi.TransactionContextInterface, binding string, timestamp int64) (string, error) {
	tx := Transaction{
		Binding:   binding,
		Timestamp: timestamp,
	}
	// compose transaction key
	txJSON, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	hash.Write(txJSON)
	txKey := hex.EncodeToString(hash.Sum(nil))
	exists, err := s.TXExists(ctx, txKey)
	if err != nil {
		return "", err
	}
	if exists {
		return "", fmt.Errorf("the transaction %s already exists", txKey)
	}
	return txKey, ctx.GetStub().PutState(txKey, txJSON)
}

// ReadTX returns the transaction stored in the world state with given id.
func (s *SmartContract) ReadTX(ctx contractapi.TransactionContextInterface, id string) (*Transaction, error) {
	txJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if txJSON == nil {
		return nil, fmt.Errorf("the transaction %s does not exist", id)
	}
	tx := new(Transaction)
	err = json.Unmarshal(txJSON, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

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
