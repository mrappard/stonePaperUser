/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (

	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"github.com/hyperledger/fabric/core/crypto/primitives"
)

// SocietyIdentifier is simple chaincode implementing a basic Asset Management system
// with access control enforcement at chaincode level.
// Look here for more information on how to implement access control at chaincode level:
// https://github.com/hyperledger/fabric/blob/master/docs/tech/application-ACL.md
// An asset is simply represented by a string.
type SocietyIdentifier struct {
}

// Init method will be called during deployment.
// The deploy transaction metadata is supposed to contain the administrator cert

func (t *SocietyIdentifier) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Init Chaincode...")
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	// Create ownership table
	err := stub.CreateTable("UserIdentity", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "User", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "Status", Type: shim.ColumnDefinition_STRING, Key: false},
	})
	if err != nil {
		return nil, errors.New("Failed creating UserIdentity table.")
	}

	fmt.Printf("Init Chaincode...done")

	return nil, nil
}

func (t *SocietyIdentifier) create(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("create...")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	user := args[0]

	ok, err := stub.InsertRow("UserIdentity", shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: user}},
			&shim.Column{Value: &shim.Column_String_{String_: "true"}}},
	})

	if !ok && err == nil {
		return nil, errors.New("Asset was already created.")
	}

	fmt.Printf("create...done!")

	return nil, err
}

func (t *SocietyIdentifier) update(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Update...")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	user := args[0]
	state := args[1]

	// Verify the identity of the caller
	// Only the owner can transfer one of his assets
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: user}}
	columns = append(columns, col1)


	err := stub.DeleteRow(
		"UserIdentity",
		[]shim.Column{shim.Column{Value: &shim.Column_String_{String_: user}}},
	)
	if err != nil {
		return nil, errors.New("Failed deleting row.")
	}

	_, err = stub.InsertRow(
		"UserIdentity",
		shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_String_{String_: user}},
				&shim.Column{Value: &shim.Column_String_{String_: state}},
			},
		})
	if err != nil {
		return nil, errors.New("Failed inserting row.")
	}


	fmt.Printf("Update...done")

	return nil, nil
}

// Invoke will be called for every transaction.
// Supported functions are the following:
// "create(asset, owner)": to create ownership of assets. An asset can be owned by a single entity.
// Only an administrator can call this function.
// "transfer(asset, newOwner)": to transfer the ownership of an asset. Only the owner of the specific
// asset can call this function.
// An asset is any string to identify it. An owner is representated by one of his ECert/TCert.
func (t *SocietyIdentifier) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)
	// Handle different functions
	if function == "create" {
		// create ownership
		return t.create(stub, args)
	} else if function == "update" {
		// Transfer ownership
		return t.update(stub, args)
	}

	return nil, errors.New("Received unknown function invocation")
}

// Query callback representing the query of a chaincode
// Supported functions are the following:
// "query(asset)": returns the owner of the asset.
// Anyone can invoke this function.
func (t *SocietyIdentifier) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Printf("Query [%s]", function)

	if function != "query" {
		return nil, errors.New("Invalid query function name. Expecting 'query' but found '" + function + "'")
	}

	var err error

	if len(args) != 1 {
		fmt.Printf("Incorrect number of arguments. Expecting name of an user to query")
		return nil, errors.New("Incorrect number of arguments. Expecting name of an user to query")
	}

	// Who is the user?
	asset := args[0]

	fmt.Printf("Arg [%s]", string(asset))

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: asset}}
	columns = append(columns, col1)

	row, err := stub.GetRow("UserIdentity", columns)
	if err != nil {
		fmt.Printf("Failed retriving user [%s]: [%s]", string(asset), err)
		return nil, fmt.Errorf("Failed retriving user [%s]: [%s]", string(asset), err)
	}

	fmt.Printf("Query done [% x]", row.Columns[1].GetBytes())

	return []byte("This is a Test"),nil//row.Columns[1].GetBytes(), nil
}

func main() {
	//primitives.SetSecurityLevel("SHA3", 256)
	err := shim.Start(new(SocietyIdentifier))
	if err != nil {
		fmt.Printf("Error starting SocietyIdentifier: %s", err)
	}
}
