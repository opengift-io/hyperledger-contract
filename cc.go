/*
OPENGIFT MAIN CONTRACT v0.1 01.2018
*/

package main

import (
	"fmt"
	"bytes"
	"strconv"
	"time"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("opengift")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type clientState struct {
	Balance  float64
	Projects map[string]float64
}

type projectState struct {
	Users map[string]int
}

type offer struct {
	Wallet    string
	Sum       float64
	Status    int
	Comment   string
	Timestamp int
	ClosedBy  string
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### OPENGIFT Init ###########")


	return shim.Success(nil)
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### OPENGIFT Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	//	if function == "delete" {
	//		// Deletes an entity from its state
	//		return t.delete(stub, args)
	//	}

	if function == "query" {
		// queries an entity state
		return t.query(stub, args)
	}
	if function == "move" {
		return t.move(stub, args)
	}

	if function == "add" {
		return t.add(stub, args)
	}

	if function == "pay" {
		return t.pay(stub, args)
	}

	if function == "getKey" {
		return t.getKey(stub, args)
	}

	if function == "addProject" {
		return t.addProject(stub, args)
	}

	if function == "donate" {
		return t.donate(stub, args)
	}
	if function == "closeOffer" {
		return t.closeOffer(stub, args)
	}
	if function == "setOffer" {
		return t.setOffer(stub, args)
	}
	if function == "getOffers" {
		return t.getOffers(stub, args)
	}
	//	if function == "firstPay" {
	//		return t.firstPay(stub, args)
	//	}

	strError := "Unknown action, check the first argument, must be one of 'delete', 'query', 'pay', 'getKey', 'donate', or 'move'. But got: %v"
	logger.Errorf(strError, args[0])
	return shim.Error(fmt.Sprintf(strError, args[0]))
}

//func (t *SimpleChaincode) firstPay(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	value, err := stub.GetState("firstPay")
//	if err != nil {
//		return shim.Error("Failed to get state")
//	}
//
//	if value != nil {
//		return shim.Error("Already payed")
//	}
//
//	sum := 160000000
//
//	pk, err := cid.GetX509CertificatePublicKey(stub)
//
//	csvalue, err := stub.GetState(pk)
//	if err != nil {
//		return shim.Error("Failed to get state")
//	}
//
//	var cs clientState
//	err = json.Unmarshal(csvalue, &cs)
//	cs.Balance = float64(sum)
//
//	err = stub.PutState("firstPay", []byte("1"))
//
//	if err != nil {
//		return shim.Error("Failed to put state of first pay")
//	}
//
//	strUser2State, err := json.Marshal(cs)
//	err = stub.PutState(pk, []byte(strUser2State))
//
//	if err != nil {
//		return shim.Error("Failed to put state of client")
//	}
//
//	return shim.Success([]byte("ok"))
//}
func transfer(stateFrom clientState, stateTo clientState, sum float64, stub shim.ChaincodeStubInterface) bool {
	if (stateFrom.Balance < sum) {
		return false
	}

	stateFrom.Balance = stateFrom.Balance - sum
	stateTo.Balance = stateTo.Balance + sum

	return true
}

func (t *SimpleChaincode) getOffers(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	startKey := "OFFER_000000000000000000000000"
	endKey := "OFFER_xxxxxxxxxxxxxxxxxxxx"
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		buffer.WriteString(queryResponse.Key + ":" + string(queryResponse.Value))
	}

	return shim.Success(buffer.Bytes())
}

func (t *SimpleChaincode) closeOffer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	pk, err := cid.GetX509CertificatePublicKey(stub)

	myKey := "OFFER_" + pk

	offerState, err := stub.GetState(myKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	if offerState == nil {
		return shim.Error("Offer not found")
	}

	var offerObject offer
	err = json.Unmarshal(offerState, &offerObject)
	if err != nil {
		return shim.Error(err.Error())
	}

	if offerObject.Status != 1 {
		return shim.Error("Offer is already paid")
	}

	err = stub.DelState(myKey)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("removed"))
}

func (t *SimpleChaincode) setOffer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	pk, err := cid.GetX509CertificatePublicKey(stub)
	var value float64
	value, err = strconv.ParseFloat(args[0], 64)
	if value <= 0 {
		return shim.Error("Failed to set value")
	}

	key := "OFFER_" + pk
	csvalue, err := stub.GetState(pk)
	if err != nil {
		return shim.Error("Failed to get state")
	}

	if csvalue == nil {
		return shim.Error("Entity not found")
	}

	var cs clientState
	err = json.Unmarshal(csvalue, &cs)
	if cs.Balance < 10000 {
		return shim.Error("Your balance must be more then 10 000 GIFTS")
	}

	comment := args[1]
	timestamp, err := strconv.Atoi(args[2])
	status := 1

	if int64(timestamp) > (time.Now().Unix() + int64(3600*2)) {
		return shim.Error("Not now")
	}

	myOffer := offer{
		Wallet:    pk,
		Sum:       value,
		Status:    status,
		Comment:   comment,
		Timestamp: timestamp,
		ClosedBy:  ""}

	strOffer2State, err := json.Marshal(myOffer)
	err = stub.PutState(key, []byte(strOffer2State))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("ok"))
}

func (t *SimpleChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var B string
	var X int
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 4, function followed by 2 names and 1 value")
	}

	A, err := cid.GetX509CertificatePublicKey(stub)
	pName := args[0]
	B = args[1]

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	var c2s clientState
	err = json.Unmarshal(Bvalbytes, &c2s)

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	pState, err := stub.GetState(pName)
	if err != nil {
		return shim.Error("Failed to get project state")
	}
	var oPState projectState
	if pState == nil {
		return shim.Error("No such project")
	}
	err = json.Unmarshal(pState, &oPState)

	if oPState.Users[A] <= 0 {
		return shim.Error("Zero balance of this project")
	}

	oPState.Users[A] = oPState.Users[A] - X
	oPState.Users[B] = oPState.Users[B] + X

	if oPState.Users[A] < 0 || oPState.Users[B] < 0 {
		return shim.Error("The balance is over")
	}

	if c2s.Projects[pName] == 0 {
		c2s.Projects[pName] = 1
		strUser2State, err := json.Marshal(c2s)
		err = stub.PutState(B, []byte(strUser2State))
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	proj2State, err := json.Marshal(oPState)
	err = stub.PutState(pName, []byte(proj2State))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(A));
}

func (t *SimpleChaincode) pay(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// must be an invoke
	var To string // Entities
	var X float64 // Transaction value
	var err error
	var stateFrom clientState
	var stateTo clientState
	var offerCode string

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting minimum 2")
	}

	pk, err := cid.GetX509CertificatePublicKey(stub)
	To = args[0]

	if To == pk {
		return shim.Error("You cant pay yourself")
	}

	X, err = strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}

	if len(args) == 3 {
		offerCode = args[2]
	}

	strAccountTo, err := stub.GetState(To)
	if err != nil {
		return shim.Error("Failed to get user 2 state")
	}

	if strAccountTo == nil {
		return shim.Error("Receiver not found")
	}

	err = json.Unmarshal(strAccountTo, &stateTo)

	strAccountFrom, err := stub.GetState(pk)
	if err != nil {
		return shim.Error("Failed to get sender state")
	}

	if strAccountFrom == nil {
		return shim.Error("Sender wallet not found")
	}

	err = json.Unmarshal(strAccountFrom, &stateFrom)
	if stateFrom.Balance <= 0 {
		return shim.Error("Zero sender balance")
	}

	if offerCode != "" {
		offerState, err := stub.GetState(offerCode)
		if err != nil {
			return shim.Error(err.Error())
		}

		if offerState == nil {
			return shim.Error("Offer not found")
		}

		var offerObject offer
		err = json.Unmarshal(offerState, &offerObject)
		if err != nil {
			return shim.Error(err.Error())
		}

		if offerObject.Sum < X {
			return shim.Error("Offer sum is lesser than payment")
		}

		offerObject.Sum = offerObject.Sum - X
		if offerObject.Sum == 0 {
			offerObject.Status = 0
			offerObject.ClosedBy = pk
		}

		offerState, err = json.Marshal(offerObject)
		err = stub.PutState(offerCode, []byte(offerState))
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	if transfer(stateFrom, stateTo, X, stub) == false {
		return shim.Error("Transfer error")
	}

	stateFromStr, err := json.Marshal(stateFrom)
	err = stub.PutState(pk, []byte(stateFromStr))
	if err != nil {
		return shim.Error(err.Error())
	}

	stateToStr, err := json.Marshal(stateTo)
	err = stub.PutState(To, []byte(stateToStr))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("ok"));
}

func (t *SimpleChaincode) addProject(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	pk, err := cid.GetX509CertificatePublicKey(stub)

	pName := args[0]
	if pName == "" {
		return shim.Error("You need to specify the Project Name")
	}

	Pvalbytes, err := stub.GetState(pName)
	if Pvalbytes != nil {
		jsonResp := "{\"Error\":\"Project already exists\"}"
		return shim.Error(jsonResp)
	}

	Avalbytes, err := stub.GetState(pk)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + pk + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"You need to add wallet at first " + pk + "\"}"
		return shim.Error(jsonResp)
	}
	users := make(map[string]int)
	users[pk] = 100

	state := projectState{
		Users: users,
	}

	strState, er := json.Marshal(state)
	if er != nil {
		return shim.Error("Failed to marshal Project state")
	}

	var cs clientState
	err = json.Unmarshal(Avalbytes, &cs)

	if cs.Projects == nil {
		cs.Projects = map[string]float64{}
	}

	cs.Projects[pName] = 1
	strUserState, er := json.Marshal(cs)
	if er != nil {
		return shim.Error("Failed to marshal User state")
	}

	err = stub.PutState(pName, []byte(strState))
	if err != nil {
		return shim.Error("Failed to add Projectstate")
	}

	err = stub.PutState(pk, []byte(strUserState))
	if err != nil {
		return shim.Error("Failed to add User state")
	}

	return shim.Success([]byte(pk))
}

func (t *SimpleChaincode) donate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	X, err := strconv.ParseFloat(args[1], 64)

	project := args[0]

	pk, err := cid.GetX509CertificatePublicKey(stub)

	Avalbytes, err := stub.GetState(pk)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + pk + "\"}"
		return shim.Error(jsonResp)
	}

	var cs clientState
	err = json.Unmarshal(Avalbytes, &cs)

	if cs.Balance < X {
		return shim.Error("balance is over")
	}

	cs.Balance = cs.Balance - X

	pState, err := stub.GetState(project)
	if err != nil {
		return shim.Error("Failed to get project state")
	}
	var oPState projectState
	if pState == nil {
		return shim.Error("No such project")
	}

	err = json.Unmarshal(pState, &oPState)
	updatedCurrent := 0
	for key, value := range oPState.Users {
		if value != 0 {
		}
		if key == pk && value == 100 {
			return shim.Error("Failed to donate yourself")
		}

		Avalbytes, err := stub.GetState(key)
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to get state for user " + pk + "\"}"
			return shim.Error(jsonResp)
		}

		var csP clientState
		err = json.Unmarshal(Avalbytes, &csP)

		csP.Balance = csP.Balance + (X * float64(oPState.Users[key]) / 100.0)
		if key == pk {
			csP.Balance = csP.Balance - X
			updatedCurrent = 1
		}

		strStateNew, er := json.Marshal(&csP)
		if er != nil {
			return shim.Error("Failed to marshal state")
		}
		er = stub.PutState(key, []byte(strStateNew))
		if er != nil {
			return shim.Error("Failed to add state")
		}
	}

	if updatedCurrent == 0 {
		strStateNew, er := json.Marshal(&cs)
		if er != nil {
			return shim.Error("Failed to marshal state")
		}
		er = stub.PutState(pk, []byte(strStateNew))
		if er != nil {
			return shim.Error("Failed to add state")
		}
	}

	return shim.Success([]byte(pk))
}

func (t *SimpleChaincode) getKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	pk, err := cid.GetX509CertificatePublicKey(stub)
	if err != nil {
		return shim.Error("fail")
	}
	return shim.Success([]byte(pk))
}

func (t *SimpleChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	pk, err := cid.GetX509CertificatePublicKey(stub)

	pr := map[string]float64{}

	state := clientState{Balance: 0.0, Projects: pr}

	Avalbytes, err := stub.GetState(pk)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + pk + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {

		strState, er := json.Marshal(&state)
		if er != nil {
			return shim.Error("Failed to marshal state")
		}

		err := stub.PutState(pk, []byte(strState))
		if err != nil {
			return shim.Error("Failed to add state")
		}

		return shim.Success([]byte(pk))
	}

	jsonResp := "{\"Error\":\"Not nil amount for " + pk + "\"}"
	return shim.Error(jsonResp)
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	logger.Infof("Query Response:%s\n", jsonResp)
	//	id, err := cid.GetID(stub)

	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
