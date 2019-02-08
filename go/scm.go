package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the car structure, with 4 properties.  Structure tags are used by encoding/json library
type Invoice struct {
	/* Make   string `json:"make"`
	Model  string `json:"model"`
	Colour string `json:"colour"`
	Owner  string `json:"owner"` */
	InvoiceNumber   string `json:"invoiceNumber"`
	BilledTo        string `json:"billedTo"`
	InvoiceDate     string `json:"invoiceDate"`
	InvoiceAmount   string `json:"invoiceAmount"`
	ItemDescription string `json:"itemDescription"`
	GoodsReceived   string `json:"gr"`
	IsPaid          string `json:"isPaid"`
	PaidAmount      string `json:"paidAmount"`
	Repaid          string `json:"repaid"`
	RepaymentAmount string `json:"repaymentAmount"`
}

/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "queryAllInvoice" {
		return s.queryAllInvoice(APIstub)
	} else if function == "receiveGoods" {
		return s.receiveGoods(APIstub, args)
	} else if function == "getHistoryForCar" {
		return s.getHistoryForCar(APIstub, args)
	} else if function == "raiseInvoice" {
		return s.raiseInvoice(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}
func (s *SmartContract) raiseInvoice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var invoices = Invoice{InvoiceNumber: args[0], BilledTo: args[1], InvoiceDate: args[2], InvoiceAmount: args[3], ItemDescription: args[4], GoodsReceived: args[5], IsPaid: args[6], PaidAmount: args[7], Repaid: args[8], RepaymentAmount: args[9]}

	invoiceAsBytes, _ := json.Marshal(invoices)
	APIstub.PutState(args[0], invoiceAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	invoices := []Invoice{
		Invoice{InvoiceNumber: "123", BilledTo: "Lenovo", InvoiceDate: "2/7/2019", InvoiceAmount: "2000", ItemDescription: "some here", GoodsReceived: "yes", IsPaid: "no", PaidAmount: "1000", Repaid: "no", RepaymentAmount: "200"},
	}

	i := 0
	for i < len(invoices) {
		fmt.Println("i is ", i)
		invoiceAsBytes, _ := json.Marshal(invoices[i])
		APIstub.PutState("INV"+strconv.Itoa(i), invoiceAsBytes)
		fmt.Println("Added", invoices[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) queryAllInvoice(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "INV0"
	endKey := "INV999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}\n")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- displayAllInvoices:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) receiveGoods(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	invoiceAsBytes, _ := APIstub.GetState(args[0])
	invoice := Invoice{}

	json.Unmarshal(invoiceAsBytes, &invoice)
	invoice.GoodsReceived = args[1]

	invoiceAsBytes, _ = json.Marshal(invoice)
	APIstub.PutState(args[0], invoiceAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) getUser(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	attr := args[0]
	attrValue, _, _ := cid.GetAttributeValue(APIstub, attr)

	msp, _ := cid.GetMSPID(APIstub)

	var buffer bytes.Buffer
	buffer.WriteString("{\"User\":")
	buffer.WriteString("\"")
	buffer.WriteString(attrValue)
	buffer.WriteString("\"")

	buffer.WriteString(", \"MSP\":")
	buffer.WriteString("\"")

	buffer.WriteString(msp + "_DUMMY_change")
	buffer.WriteString("\"")

	buffer.WriteString("}")

	return shim.Success(buffer.Bytes())

	//return shim.Success(nil)
}

func (s *SmartContract) getHistoryForCar(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carKey := args[0]

	resultsIterator, err := APIstub.GetHistoryForKey(carKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the car
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		buffer.WriteString(string(response.Value))

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
