package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/inklabsfoundation/inkchain/core/chaincode/shim"
	pb "github.com/inklabsfoundation/inkchain/protos/peer"
)

// Incentive-related const
const (
	IncentiveBalanceType  = "INK"
	IncentiveMashupInvoke = "10"
)

// Definitions of a service's status
const (
	S_Created   = "created"
	S_Available = "available"
	S_Invalid   = "invalid"
)

// Prefixes for user and service separately
const (
	UserPrefix    = "USER_"
	ServicePrefix = "SER_"
)

// Invoke functions definition
const (
	// User-related basic invoke
	RegisterUser = "registerUser"
	RemoveUser   = "removeUser"
	QueryUser    = "queryUser"

	// Service-related invoke
	RegisterService     = "registerService"
	InitAccount         = "initAccount"
	InvalidateService   = "invalidateService" // mark whether the service is validated
	PublishService      = "publishService"    // publish a created service
	CreateMashup        = "createMashup"      // utilize services to create a new mashup
	QueryService        = "queryService"
	EditService         = "editService"
	QueryServiceByUser  = "queryServiceByUser"
	QueryServiceByRange = "queryServiceByRange"
	GivesToken          = "givesToken"
	InvokeService       = "invokeService"

	// User-related reward invoke
	RewardService = "rewardService"

	Created    string = "created"
	Delivered  string = "issued"
	Invalidate string = "invalidated"
)

// Chaincode for DSES (Decentralized Service Eco-System)
type serviceChaincode struct {
}

// Structure definition for user
type user struct {
	Name         string `json:"name"`
	Introduction string `json:"introduction"`
	Address      string `json:"address"`
	// There is a one-to-one correspondence between "Name" and "Address"
	// The Address records the user's profit from creating valuable services or mashups.

	Contribution   int `json:"contribution"`
	DeveloperToken int `json:"developerToken"`
	// "Contribution" evaluates the user's contribution to the service ecosystem.
	// TODO: add handler about "Contribution"
	// Benefit of "Contribution":
	// 1. construct a evaluation for every user's contribution on the service ecosystem
	// 2. inspire users to participate in creating new services and mashups

}

// type GenAccount
type Token struct {
	// token name
	Name string `json:"tokenName"`
	// total supply of the token
	totalSupply *big.Int `json:"totalSupply"`
	// initial address to issue
	Address string `json:"address"`
	// token status : Created, Delivered, Invalidate
	Status string `json:"status"`
	// token decimals
	Decimals int `json:"decimals"`
}

// Structure definition for service
// type "service" defines conventional services as well as mashups.
type service struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Developer   string `json:"developer"` // record the user that developed this service
	Description string `json:"description"`

	CreatedTime string `json:"createdTime"`
	UpdatedTime string `json:"updatedTime"`

	// Status records the status of a service:
	// created/available/invalid
	Status string `json:"status"`

	// Whether the service is a mashup or not.
	IsMashup bool `json:"isMashup"`

	// if the service is a mashup, "Composited" records the services that it invokes;
	// if the service is not a mashup, "Composited" records the co-occurrence documents of the service
	Composition map[string]int `json:"composition"`

	// Benefit of "Composited":
	// 1. Automatically create service co-occurrence documents and store it into the ledger
	// 2. Promote the security and integrality of service data

	// future: people need to pay if they want to use the record information
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(serviceChaincode))
	if err != nil {
		fmt.Printf("Error starting assetChaincode: %s", err)
	}
}

// Init initializes chaincode
// ==================================================================================
func (t *serviceChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("assetChaincode Init.")
	return shim.Success([]byte("Init success."))
}

// Invoke func
// ==================================================================================
func (t *serviceChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("assetChaincode Invoke.")
	function, args := stub.GetFunctionAndParameters()

	switch function {
	// ********************************************************
	// PART 1: User-related invokes
	case RegisterUser:
		if len(args) != 2 {
			return shim.Error("Incorrect number of arguments. Expecting 2.")
		}
		// args[0]: user name
		return t.registerUser(stub, args)

	case RemoveUser:
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments. Expecting 1.")
		}
		// args[0]: user name
		return t.removeUser(stub, args)

	case QueryUser:
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments. Expecting 1.")
		}
		// args[0]: user name
		return t.queryUser(stub, args)

	case InitAccount:
		if len(args) != 4 {
			return shim.Error("Incorrect number of arguments. Expecting 4.")
		}
		// args[0]: user name
		return t.initAccount(stub, args)

	// ********************************************************
	// PART 2: service-related invokes
	case RegisterService:
		if len(args) != 4 {
			return shim.Error("Incorrect number of arguments. Expecting 4.")
		}
		// args[0]: service name
		// args[1]: service type
		// args[2]: service description
		// args[3]: developer's name
		return t.registerService(stub, args)

	case InvalidateService:
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments. Expecting 1.")
		}
		// args[0]: service name
		return t.invalidateService(stub, args)

	case PublishService:
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments. Expecting 1.")
		}
		// args[0]: service name
		return t.publishService(stub, args)

	case QueryService:
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments. Expecting 1.")
		}
		// args[0]: service name
		return t.queryService(stub, args)

	case EditService:
		if len(args) != 3 {
			return shim.Error("Incorrect number of arguments. Expecting 3.")
		}
		// args[0]: service name
		// args[1]: filed name to change
		// args[2]: new filed value
		return t.editService(stub, args)

	case CreateMashup:
		if len(args) < 4 {
			return shim.Error("Incorrect number of arguments. Expecting 4 at least.")
		}
		// args[0]: mashup name
		// args[1]: mashup type
		// args[2]: mashup description
		// args[3...]: invoked service list
		return t.createMashup(stub, args)

	case QueryServiceByRange:
		if len(args) != 2 {
			return shim.Error("Incorrect number of arguments. Expecting 2.")
		}
		// args[0]: begin index
		// args[1]: end index
		return t.queryServiceByRange(stub, args)

	// ********************************************************
	// PART 3: user-related reward invokes
	case RewardService:
		if len(args) < 3 {
			return shim.Error("Incorrect number of arguments. Expecting 3 at least.")
		}
		// args[0]: service name
		// args[1]: reward_type
		// args[2]: reward_amount
		return t.rewardService(stub, args)

	case GivesToken:
		if len(args) < 2 {
			return shim.Error("Incorrect number of arguments. Expecting 2 at least.")
		}
		// args[0]: service name
		// args[1]: reward_type
		// args[2]: reward_amount
		return t.givesToken(stub, args)

	case InvokeService:
		if len(args) < 2 {
			return shim.Error("Incorrect number of arguments. Expecting 2 at least.")
		}
		// args[0]: service name
		// args[1]: reward_type
		// args[2]: reward_amount
		return t.invokeService(stub, args)
	}

	return shim.Error("Invalid invoke function name.")
}

// Invoke func about user
// ==================================================================================

// ==================================
// registerUser: Register a new user
// ==================================
func (t *serviceChaincode) registerUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var new_name string
	var new_intro string
	var new_add string
	var err error

	new_name = args[0]
	new_intro = args[1]

	// Get the user's address automatically through INKchian's GetSender() interface
	new_add, err = stub.GetSender()
	if err != nil {
		return shim.Error("Fail to get the sender's address.")
	}

	// check if user exists
	user_key := UserPrefix + new_name
	userAsBytes, err := stub.GetState(user_key)
	if err != nil {
		return shim.Error("Fail to get user: " + err.Error())
	} else if userAsBytes != nil {
		return shim.Error("This user already exists: " + new_name)
	}

	// register user
	user := &user{new_name, new_intro, new_add, 0, 0}
	userJSONasBytes, err := json.Marshal(user)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(user_key, userJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("User register & Init account success."))
}

// ==================================
// initAccount: Initate token for new user accounr
// ==================================
func (t *serviceChaincode) initAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// var A string           // Address
	// var BalanceType string // Token type

	// var err error

	// tokenName := args[1]

	// A = strings.ToLower(args[0])
	// BalanceType = args[1]

	// source_add, err = stub.GetSender()
	// if err != nil {
	// 	return shim.Error("Fail to get the sender's address.")
	// }

	// // Get the state from the ledger
	// account, err := stub.GetAccount(A)
	// if err != nil {
	// 	jsonResp := "{\"Error\":\"account not exists\"}"
	// 	return shim.Error(jsonResp)
	// }

	// if account == nil {
	// 	jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
	// 	return shim.Error(jsonResp)
	// }
	// balanceJson, jsonErr := json.Marshal(account.Balance)
	// if jsonErr != nil {
	// 	return shim.Error(jsonErr.Error())
	// }

	// // Amount
	// amount := big.NewInt(0)
	// _, good := amount.SetString("100", 10)
	// if !good {
	// 	return shim.Error("Expecting integer value for amount")
	// }

	// //Get exist token
	// var existToken Token
	// existTokenBytes, err := stub.GetState(tokenName)
	// if err != nil {
	// 	msgCheck := "Check token existance error, fail to getState of "
	// 	msgCheck += tokenName
	// 	// tralogger.Debug(msgCheck)
	// 	return shim.Error(msgCheck)
	// }

	// //Get the information of token
	// //If not exist, create a new token first
	// if existTokenBytes == nil {
	// 	//not exist
	// 	//create the token
	// 	existToken.Status = Created
	// 	existToken.Name = tokenName
	// 	existToken.totalSupply = amount
	// 	existToken.Address = A
	// 	existToken.Decimals = 10
	// } else {
	// 	//exist
	// 	//unmarshal the jsonBytes & check token information
	// 	err = json.Unmarshal(existTokenBytes, &existToken)
	// 	if err != nil {
	// 		msgUnmarshal := "Unmarshal exist tokenBytes err "
	// 		msgUnmarshal += tokenName
	// 		// tralogger.Debug(msgUnmarshal)
	// 		return shim.Error(msgUnmarshal)
	// 	}
	// 	//check the status of token
	// 	if existToken.Status != Created {
	// 		msgCheckTS := "Token status err, fail to issue token."
	// 		// tralogger.Debug(msgCheckTS)
	// 		return shim.Error(msgCheckTS)
	// 	}
	// 	//check the information of token
	// 	if existToken.Address != A || existToken.totalSupply.Cmp(amount) != 0 || existToken.Decimals != 10 {
	// 		msgCheckTInfo := "Token info err, check fialed."
	// 		// tralogger.Debug(msgCheckTInfo)
	// 		return shim.Error(msgCheckTInfo)
	// 	}
	// }

	// err = stub.IssueToken(A, BalanceType, amount)
	// // if err != nil {
	// // 	return shim.Error("transfer error" + err.Error())
	// // }
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	// existToken.Status = Delivered

	// //store the latest status for token in ascc
	// existTokenJson, err := json.Marshal(&existToken)
	// err = stub.PutState(tokenName, existTokenJson)

	// if err != nil {
	// 	msgUpdate := "Store the latest token status err."
	// 	// tralogger.Debug(msgUpdate)
	// 	return shim.Error(msgUpdate)
	// }

	// jsonResp := "{\"Name\":\"" + A + "\",\"Balance\":\"" + string(balanceJson[:]) + "\"}"
	// return shim.Success([]byte(jsonResp))
	var err error

	tokenName := args[0]

	totalSupply := big.NewInt(0)
	_, good := totalSupply.SetString(args[1], 10)
	if !good {
		return shim.Error("Expecting integer value for totalSupply.")
	}
	dec, _ := strconv.Atoi(args[2])
	addr := args[3]

	//Get exist token
	var existToken Token
	// var existTokenBytes []byte
	existTokenBytes, err := stub.GetState(tokenName)
	if err != nil {
		msgCheck := "Check token existance error, fail to getState of "
		msgCheck += tokenName
		// tralogger.Debug(msgCheck)
		return shim.Error(msgCheck)
	}

	//Get the information of token
	//If not exist, create a new token first
	if existTokenBytes == nil {
		//not exist
		//create the token
		existToken.Status = Created
		existToken.Name = tokenName
		existToken.totalSupply = totalSupply
		existToken.Address = addr
		existToken.Decimals = dec
	} else {
		//exist
		//unmarshal the jsonBytes & check token information
		err = json.Unmarshal(existTokenBytes, &existToken)
		if err != nil {
			msgUnmarshal := "Unmarshal exist tokenBytes err "
			msgUnmarshal += tokenName
			// tralogger.Debug(msgUnmarshal)
			return shim.Error(msgUnmarshal)
		}
		//check the status of token
		if existToken.Status != Created {
			msgCheckTS := "Token status err, fail to issue token."
			// tralogger.Debug(msgCheckTS)
			return shim.Error(msgCheckTS)
		}
		//check the information of token
		// || existToken.Decimals != dec
		if existToken.Address != addr || existToken.totalSupply.Cmp(totalSupply) != 0 {
			msgCheckTInfo := "Token info err, check fialed."
			// tralogger.Debug(msgCheckTInfo)
			return shim.Error(msgCheckTInfo)
		}
	}

	//set the token number to address

	//get account of Address

	// account, err := stub.GetAccount(addr)
	// //check if token has been issued before
	// if err == nil {
	// 	if account != nil {
	// 		if _, ok := account.Balance[tokenName]; ok {
	// 			msgBalanceCheck := "Token " + tokenName + " already exist in " + addr
	// 			// tralogger.Debug(msgBalanceCheck)
	// 			return shim.Error(msgBalanceCheck)
	// 		}
	// 	}
	// }
	//token hasnot been issued, then
	//issue token
	err = stub.Transfer(addr, tokenName, totalSupply)
	if err != nil {
		return shim.Error("DSES" + err.Error())
	}

	// existToken.Status = Delivered

	//store the latest status for token in ascc
	existTokenJson, err := json.Marshal(&existToken)
	err = stub.PutState(tokenName, existTokenJson)

	if err != nil {
		msgUpdate := "Store the latest token status err."
		// tralogger.Debug(msgUpdate)
		return shim.Error(msgUpdate)
	}
	// jsonResp := "{\"Name\":\"" + A + "\",\"Balance\":\"" + string(balanceJson[:]) + "\"}"
	// return shim.Success([]byte(jsonResp))
	return shim.Success([]byte("Token issued success!"))

}

// ===================================
// removeUser: Remove an existed user
// ===================================
func (t *serviceChaincode) removeUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var user_name string
	var err error

	user_name = args[0]

	// check if user exists
	user_key := UserPrefix + user_name
	userAsBytes, err := stub.GetState(user_key)
	if err != nil {
		return shim.Error("Fail to get user: " + err.Error())
	} else if userAsBytes == nil {
		return shim.Error("This user does not exist: " + user_name)
	}

	err = stub.DelState(user_key)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("User delete success."))
}

// ===================================
// queryUser: Query an existed user
// ===================================
func (t *serviceChaincode) queryUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var user_name string
	var err error

	user_name = args[0]

	// check if user exists
	user_key := UserPrefix + user_name
	userAsBytes, err := stub.GetState(user_key)
	if err != nil {
		return shim.Error("Fail to get user: " + err.Error())
	} else if userAsBytes == nil {
		return shim.Error("This user does not exist: " + user_name)
	}

	// return user info
	return shim.Success(userAsBytes)
}

// Invoke func about service
// ==================================================================================

// =======================================
// registerService: Register a new service
// =======================================
func (t *serviceChaincode) registerService(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var service_name string
	var service_type string
	var service_des string
	var service_dev string
	var user_name string
	var err error

	service_name = args[0]
	service_type = args[1]
	service_des = args[2]
	user_name = args[3]

	// get service developer, check if it corresponds with the input user
	service_dev, err = stub.GetSender()
	if err != nil {
		return shim.Error("Fail to get the sender's address.")
	}
	user_key := UserPrefix + user_name
	userAsBytes, err := stub.GetState(user_key)
	if err != nil {
		return shim.Error("Fail to get user: " + err.Error())
	}
	var userJSON user
	err = json.Unmarshal([]byte(userAsBytes), &userJSON)
	if err != nil {
		return shim.Error("Error unmarshal user bytes.")
	}
	if userJSON.Address != service_dev {
		return shim.Error("Not the correct user.")
	}

	// update developerToken user
	newtoken := userJSON.DeveloperToken + 1
	user := &user{userJSON.Name, userJSON.Introduction, userJSON.Address, userJSON.Contribution, newtoken}
	userJSONasBytes, err := json.Marshal(user)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(user_key, userJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// check if service exists
	service_key := ServicePrefix + service_name
	serviceAsBytes, err := stub.GetState(service_key)
	if err != nil {
		return shim.Error("Fail to get service: " + err.Error())
	} else if serviceAsBytes != nil {
		return shim.Error("This service already exists: " + service_name)
	}

	// get current time
	tNow := time.Now()
	tString := tNow.UTC().Format(time.UnixDate)

	// register service
	newS := &service{service_name, service_type, user_name,
		service_des, tString, "", S_Created,
		false, make(map[string]int)}
	serviceJSONasBytes, err := json.Marshal(newS)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(service_key, serviceJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// result := givesToken(stub, user_name, "INK", "100")
	// if result != "Ok" {
	// 	return shim.Error("err.Error()")
	// }

	return shim.Success([]byte("Service register success."))
}

// =================================================
// invalidateService: Invalidate an existed service
// =================================================
func (t *serviceChaincode) invalidateService(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var service_name string
	var err error

	service_name = args[0]

	// STEP 0: check if service exists
	service_key := ServicePrefix + service_name
	serviceAsBytes, err := stub.GetState(service_key)
	if err != nil {
		return shim.Error("Fail to get service: " + err.Error())
	} else if serviceAsBytes == nil {
		return shim.Error("This service does not exists: " + service_name)
	}

	// STEP 1: check whether it is the service's developer's invocation
	var senderAdd string
	senderAdd, err = stub.GetSender()
	if err != nil {
		return shim.Error("Fail to get the sender's address.")
	}

	var serviceJSON service
	err = json.Unmarshal([]byte(serviceAsBytes), &serviceJSON)
	if err != nil {
		return shim.Error("Error unmarshal service bytes.")
	}

	// 0125
	// get developer's address
	dev_key := UserPrefix + serviceJSON.Developer
	devAsBytes, err := stub.GetState(dev_key)
	if err != nil {
		return shim.Error("Error get the developer.")
	}
	var DevJSON user
	err = json.Unmarshal([]byte(devAsBytes), &DevJSON)

	fmt.Println("DevAddress:  " + DevJSON.Address)
	if senderAdd != DevJSON.Address {
		return shim.Error("Aurthority err! Not invoke by the service's developer.")
	}

	// STEP 2: invalidate the service and store it.
	// new service, make it invalidated
	new_service := &service{serviceJSON.Name, serviceJSON.Type, serviceJSON.Developer,
		serviceJSON.Description, serviceJSON.CreatedTime, serviceJSON.UpdatedTime,
		S_Invalid, serviceJSON.IsMashup, serviceJSON.Composition}
	// store the new service
	assetJSONasBytes, err := json.Marshal(new_service)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(service_key, assetJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Invalidate Service success."))
}

// =================================================
// publishService: publish a created service
// =================================================
func (t *serviceChaincode) publishService(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var service_name string
	var err error

	service_name = args[0]

	// STEP 0: check if service exists
	service_key := ServicePrefix + service_name
	serviceAsBytes, err := stub.GetState(service_key)
	if err != nil {
		return shim.Error("Fail to get service: " + err.Error())
	} else if serviceAsBytes == nil {
		return shim.Error("This service does not exists: " + service_name)
	}

	// STEP 1: check whether it is the service's developer's invocation
	var senderAdd string
	senderAdd, err = stub.GetSender()
	if err != nil {
		return shim.Error("Fail to get the sender's address.")
	}

	var serviceJSON service
	err = json.Unmarshal([]byte(serviceAsBytes), &serviceJSON)
	if err != nil {
		return shim.Error("Error unmarshal service bytes.")
	}

	fmt.Println("SenderAdd:  " + senderAdd)
	fmt.Println("Developer:  " + serviceJSON.Developer)

	// 0125
	// get developer's address
	dev_key := UserPrefix + serviceJSON.Developer
	devAsBytes, err := stub.GetState(dev_key)
	if err != nil {
		return shim.Error("Error get the developer.")
	}
	var DevJSON user
	err = json.Unmarshal([]byte(devAsBytes), &DevJSON)

	fmt.Println("DevAddress:  " + DevJSON.Address)
	if senderAdd != DevJSON.Address {
		return shim.Error("Aurthority err! Not invoke by the service's developer.")
	}

	// STEP 2: publish the service and store it.
	// new service, make it invalidated
	new_service := &service{serviceJSON.Name, serviceJSON.Type, serviceJSON.Developer,
		serviceJSON.Description, serviceJSON.CreatedTime, serviceJSON.UpdatedTime,
		S_Available, serviceJSON.IsMashup, serviceJSON.Composition}
	// store the new service
	serviceJSONasBytes, err := json.Marshal(new_service)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(service_key, serviceJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Publish Service success."))
}

// ======================================
// queryService: Query an existed service
// ======================================
func (t *serviceChaincode) queryService(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var service_name string
	var err error

	service_name = args[0]

	// check if service exists
	service_key := ServicePrefix + service_name
	serviceAsBytes, err := stub.GetState(service_key)
	if err != nil {
		return shim.Error("Fail to get service: " + err.Error())
	} else if serviceAsBytes == nil {
		return shim.Error("This service does not exist: " + service_name)
	}

	// return service info
	return shim.Success(serviceAsBytes)
}

// ======================================
// editService: Edit an existed service
// ======================================
func (t *serviceChaincode) editService(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var service_name string
	var field_name string
	var field_value string
	var err error

	service_name = args[0]
	field_name = args[1]
	field_value = args[2]

	// STEP 0: check the service does not exist
	service_key := ServicePrefix + service_name
	serviceAsBytes, err := stub.GetState(service_key)
	if err != nil {
		return shim.Error("Fail to get service: " + err.Error())
	} else if serviceAsBytes == nil {
		return shim.Error("This service does not exist: " + service_name)
	}

	// STEP 1: check whether it is the service's developer's invocation
	var senderAdd string
	senderAdd, err = stub.GetSender()
	if err != nil {
		return shim.Error("Fail to get the sender's address.")
	}

	var serviceJSON service
	err = json.Unmarshal([]byte(serviceAsBytes), &serviceJSON)
	if err != nil {
		return shim.Error("Error unmarshal service bytes.")
	}

	// 0125
	// get developer's address
	dev_key := UserPrefix + serviceJSON.Developer
	devAsBytes, err := stub.GetState(dev_key)
	if err != nil {
		return shim.Error("Error get the developer.")
	}
	var DevJSON user
	err = json.Unmarshal([]byte(devAsBytes), &DevJSON)

	fmt.Println("DevAddress:  " + DevJSON.Address)
	if senderAdd != DevJSON.Address {
		return shim.Error("Aurthority err! Not invoke by the service's developer.")
	}

	// STEP 2: update time information
	tNow := time.Now()
	tString := tNow.UTC().Format(time.UnixDate)

	new_service := &service{serviceJSON.Name, serviceJSON.Type, serviceJSON.Developer,
		serviceJSON.Description, serviceJSON.CreatedTime, tString,
		serviceJSON.Status, serviceJSON.IsMashup, serviceJSON.Composition}

	// STEP 3: update field value
	// developer can update service's type/description information
	switch field_name {
	case "Type":
		new_service.Type = field_value
		goto LABEL_STORE
	case "Description":
		new_service.Description = field_value
		goto LABEL_STORE
	}
	return shim.Error("Error field name.")

LABEL_STORE:
	// STEP 4: store the service
	serviceJSONasBytes, err := json.Marshal(new_service)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(service_key, serviceJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// return service info
	return shim.Success(serviceAsBytes)
}

// =======================================================
// createMashup: Create a new mashup
// note: a mashup should invoke at least one service API
// =======================================================
func (t *serviceChaincode) createMashup(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var mashup_name string
	var mashup_type string
	var mashup_des string
	var mashup_dev string
	var err error

	mashup_name = args[0]
	mashup_type = args[1]
	mashup_des = args[2]

	// STEP 0: get mashup developer
	mashup_dev, err = stub.GetSender()
	if err != nil {
		return shim.Error("Fail to get the sender's address.")
	}

	// STEP 1: check if service does not exist
	mashup_key := ServicePrefix + mashup_name
	serviceAsBytes, err := stub.GetState(mashup_key)
	if err != nil {
		return shim.Error("Fail to get service: " + err.Error())
	} else if serviceAsBytes != nil {
		return shim.Error("This service already exists: " + mashup_name)
	}

	// STEP 2: create a new mashup
	// get current time
	tNow := time.Now()
	tString := tNow.UTC().Format(time.UnixDate)

	// create composition
	new_map := make(map[string]int)
	new_developer_map := make(map[string]int)
	for i := 3; i < len(args); i++ {
		// check the service exist
		service_key := ServicePrefix + args[i]
		serviceAsBytes, err := stub.GetState(service_key)
		if err != nil {
			return shim.Error("Fail to get service: " + err.Error())
		} else if serviceAsBytes == nil {
			return shim.Error("This service doesn't exist: " + args[i])
		}
		// add the service into map
		new_map[args[i]] = 1
		// temporarily store their addresses
		var serviceJSON service
		err = json.Unmarshal([]byte(serviceAsBytes), &serviceJSON)
		if err != nil {
			return shim.Error("Error unmarshal service bytes.")
		}
		new_developer_map[serviceJSON.Developer] = 1
	}

	// new mashup
	newS := &service{mashup_name, mashup_type, mashup_dev,
		mashup_des, tString, "", S_Created,
		true, new_map}

	// STEP 3: pay to the invoked services' developers
	// Important!
	// Incentive Mechanism Here

	incentive_amount := big.NewInt(0)
	incentive_amount.SetString(IncentiveMashupInvoke, 10)

	for k, _ := range new_developer_map {
		// get the k's address
		user_key := UserPrefix + k
		userAsBytes, err := stub.GetState(user_key)
		if err != nil {
			return shim.Error("Fail to get user: " + err.Error())
		} else if userAsBytes == nil {
			return shim.Error("This user doesn't exist: " + k)
		}
		var userJSON user
		err = json.Unmarshal([]byte(userAsBytes), &userJSON)
		if err != nil {
			return shim.Error("Error unmarshal user bytes.")
		}
		// make incentive transfer
		// from the mashup developer to the invoked service's developer
		err = stub.Transfer(userJSON.Address, IncentiveBalanceType, incentive_amount)
		if err != nil {
			return shim.Error("Error when making transfer.")
		}

		// update developerToken user
		newtoken := userJSON.DeveloperToken + 1
		user := &user{userJSON.Name, userJSON.Introduction, userJSON.Address, userJSON.Contribution, newtoken}
		userJSONasBytes, err := json.Marshal(user)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(user_key, userJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	// STEP 4: store the new mashup
	serviceJSONasBytes, err := json.Marshal(newS)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(mashup_key, serviceJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Mashup register success."))
}

// =======================================================
// rewardService: reward a service
// reward a service's developer, transfer fixed amount of
// specific reward_type token to the developer's account.
// =======================================================
func (t *serviceChaincode) rewardService(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var service_name string
	var reward_type string
	var err error

	service_name = args[0]
	reward_type = args[1]

	// Amount
	reward_amount := big.NewInt(0)
	_, good := reward_amount.SetString(args[2], 10)
	if !good {
		return shim.Error("Expecting integer value for amount")
	}

	// STEP 0: get service's developer
	service_key := ServicePrefix + service_name
	serviceAsBytes, err := stub.GetState(service_key)
	if err != nil {
		return shim.Error("Fail to get the service's info.")
	}

	var serviceJSON service
	err = json.Unmarshal([]byte(serviceAsBytes), &serviceJSON)
	if err != nil {
		return shim.Error("Error unmarshal service bytes.")
	}

	dev := serviceJSON.Developer

	// STEP 1: get the address of the dev
	user_key := UserPrefix + dev
	userAsBytes, err := stub.GetState(user_key)
	if err != nil {
		return shim.Error("Fail to get the developer's info.")
	}
	var userJSON user
	err = json.Unmarshal([]byte(userAsBytes), &userJSON)
	if err != nil {
		return shim.Error("Error unmarshal user bytes.")
	}

	// STEP 3: reward the developer
	toAdd := userJSON.Address
	err = stub.Transfer(toAdd, reward_type, reward_amount)
	if err != nil {
		return shim.Error("Fail realize the reawrd.")
	}

	// update developerToken user
	newtoken := userJSON.DeveloperToken + 1
	user := &user{userJSON.Name, userJSON.Introduction, userJSON.Address, userJSON.Contribution, newtoken}
	userJSONasBytes, err := json.Marshal(user)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(user_key, userJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Reward the service success."))
}

// ========================================================================
// queryServiceByRange: query services' names by range (startKey, endKey)
//
// startKey and endKey are case-sensitive
// use "" for both startKey and endKey if you want to query all the assets
// ========================================================================
func (t *serviceChaincode) queryServiceByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	startKey := ""
	endKey := ""

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	bArrayIndex := 1
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		// index of the result
		buffer.WriteString("{\"Number\":")
		buffer.WriteString("\"")
		bArrayIndexStr := strconv.Itoa(bArrayIndex)
		buffer.WriteString(string(bArrayIndexStr))
		bArrayIndex += 1
		buffer.WriteString("\"")
		// information about current asset
		buffer.WriteString(", \"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true

	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())

}

// =======================================================
// givesToken: reward a service
// reward a service's developer, transfer fixed amount of
// specific reward_type token to the developer's account.
// =======================================================
func (t *serviceChaincode) invokeService(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var service_name string
	service_name = args[0]
	//get developer from service name
	service_key := ServicePrefix + service_name
	serviceAsBytes, err := stub.GetState(service_key)
	if err != nil {
		return shim.Error("Fail to get the service's info.")
	}

	var serviceJSON service
	err = json.Unmarshal([]byte(serviceAsBytes), &serviceJSON)
	if err != nil {
		return shim.Error("Error unmarshal service bytes.")
	}

	dev := serviceJSON.Developer

	// STEP 1: get the address of the dev
	user_key := UserPrefix + dev
	userAsBytes, err := stub.GetState(user_key)
	if err != nil {
		return shim.Error("Fail to get the developer's info.")
	}
	var userJSON user
	err = json.Unmarshal([]byte(userAsBytes), &userJSON)
	if err != nil {
		return shim.Error("Error unmarshal user bytes.")
	}

	// update developerToken user
	newtoken := userJSON.DeveloperToken + 2
	user := &user{userJSON.Name, userJSON.Introduction, userJSON.Address, userJSON.Contribution, newtoken}
	userJSONasBytes, err := json.Marshal(user)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(user_key, userJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Reward the service success."))
	// return "Ok"
}

// =======================================================
// givesToken: reward a service
// reward a service's developer, transfer fixed amount of
// specific reward_type token to the developer's account.
// =======================================================
func (t *serviceChaincode) givesToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var reward_type string
	var userName string
	var incentive_type string
	var amount string
	var err error

	reward_type = args[0]
	userName = args[1]
	incentive_type = args[2]

	switch incentive_type {
	// ************************ Developers token ***********************
	// register service
	case "1":
		amount = "110"
		break
	// register mashup
	case "2":
		amount = "110"
		break
	// service is invoked
	case "3":
		amount = "110"
		break
	// user gives token to service provider
	case "4":
		amount = "110"
		break

	// ************************ Users token ***********************
	// register user
	case "5":
		amount = "510"
		break
	// comments
	case "6":
		amount = "110"
		break
	// thumbps up/down (every 10)
	case "7":
		amount = "110"
		break

	}
	// Amount
	reward_amount := big.NewInt(0)
	_, good := reward_amount.SetString(amount, 10)
	if !good {
		return shim.Error("Expecting integer value for amount")
		// return "Error"
	}

	// STEP 1: get the address of the dev
	user_key := UserPrefix + userName
	userAsBytes, err := stub.GetState(user_key)
	if err != nil {
		return shim.Error("Fail to get the developer's info.")
		// return "Error"
	}
	var userJSON user
	err = json.Unmarshal([]byte(userAsBytes), &userJSON)
	if err != nil {
		return shim.Error("Error unmarshal user bytes.")
		// return "Error"
	}

	// STEP 3: reward the developer
	toAdd := userJSON.Address
	err = stub.Transfer(toAdd, reward_type, reward_amount)
	if err != nil {
		return shim.Error("Fail realize the reawrd.")
		// return "Error"
	}

	return shim.Success([]byte("Reward the service success."))
	// return "Ok"
}
