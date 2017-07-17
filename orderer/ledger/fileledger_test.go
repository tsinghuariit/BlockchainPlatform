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

package ordererledger_test

import (
	"io/ioutil"
	"os"

	"github.com/hyperledger/fabric/orderer/common/bootstrap/provisional"
	. "github.com/hyperledger/fabric/orderer/ledger"
	fileledger "github.com/hyperledger/fabric/orderer/ledger/file"
	"github.com/hyperledger/fabric/orderer/localconfig"
	cb "github.com/hyperledger/fabric/protos/common"
)

var genesisBlock *cb.Block

func init() {
	genesisBlock = provisional.New(config.Load()).GenesisBlock()
	testables = append(testables, &fileLedgerTestEnv{})
}

type fileLedgerTestFactory struct {
	location string
}

type fileLedgerTestEnv struct {
}

func (env *fileLedgerTestEnv) Initialize() (ledgerTestFactory, error) {
	var err error
	location, err := ioutil.TempDir("", "hyperledger")
	if err != nil {
		return nil, err
	}
	return &fileLedgerTestFactory{location: location}, nil
}

func (env *fileLedgerTestEnv) Name() string {
	return "fileledger"
}

func (env *fileLedgerTestFactory) Destroy() error {
	err := os.RemoveAll(env.location)
	return err
}

func (env *fileLedgerTestFactory) Persistent() bool {
	return true
}

func (env *fileLedgerTestFactory) New() (Factory, ReadWriter) {
	flf := fileledger.New(env.location)
	fl, err := flf.GetOrCreate(provisional.TestChainID)
	if err != nil {
		panic(err)
	}
	fl.Append(genesisBlock)
	return flf, fl
}
