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

package statebasedval

import (
	"testing"

	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb/stateleveldb"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/version"
	"github.com/hyperledger/fabric/core/ledger/testutil"
)

func TestCombinedIterator(t *testing.T) {
	testDBEnv := stateleveldb.NewTestVDBEnv(t)
	defer testDBEnv.Cleanup()

	db, err := testDBEnv.DBProvider.GetDBHandle("TestDB")
	testutil.AssertNoError(t, err, "")

	//populate db with initial data
	batch := statedb.NewUpdateBatch()
	batch.Put("ns", "key1", []byte("value1"), version.NewHeight(1, 1))
	batch.Put("ns", "key4", []byte("value4"), version.NewHeight(1, 1))
	batch.Put("ns", "key6", []byte("value6"), version.NewHeight(1, 1))
	db.ApplyUpdates(batch, version.NewHeight(1, 5))

	// prepare batch1
	batch1 := statedb.NewUpdateBatch()
	batch1.Put("ns", "key3", []byte("value3"), version.NewHeight(1, 1))
	batch1.Delete("ns", "key5", version.NewHeight(1, 1))
	batch1.Put("ns", "key6", []byte("value6_new"), version.NewHeight(1, 1))
	batch1.Put("ns", "key7", []byte("value7"), version.NewHeight(1, 1))

	// prepare batch2 (empty)
	batch2 := statedb.NewUpdateBatch()

	// Test db + batch1 updates
	dbItr1, _ := db.GetStateRangeScanIterator("ns", "key2", "key8")
	itr1, _ := newCombinedIterator("ns", dbItr1, batch1.GetRangeScanIterator("ns", "key2", "key8"))
	defer itr1.Close()

	checkItrResults(t, itr1, []*statedb.VersionedKV{
		constructVersionedKV("ns", "key3", []byte("value3"), version.NewHeight(1, 1)),
		constructVersionedKV("ns", "key4", []byte("value4"), version.NewHeight(1, 1)),
		constructVersionedKV("ns", "key6", []byte("value6_new"), version.NewHeight(1, 1)),
		constructVersionedKV("ns", "key7", []byte("value7"), version.NewHeight(1, 1)),
	})

	// Test db + batch2 updates
	dbItr2, _ := db.GetStateRangeScanIterator("ns", "key2", "key8")
	itr2, _ := newCombinedIterator("ns", dbItr2, batch2.GetRangeScanIterator("ns", "key2", "key8"))
	defer itr2.Close()
	checkItrResults(t, itr2, []*statedb.VersionedKV{
		constructVersionedKV("ns", "key4", []byte("value4"), version.NewHeight(1, 1)),
		constructVersionedKV("ns", "key6", []byte("value6"), version.NewHeight(1, 1)),
	})

	// Test db + batch1 updates with full range query
	dbItr3, _ := db.GetStateRangeScanIterator("ns", "", "")
	itr3, _ := newCombinedIterator("ns", dbItr3, batch1.GetRangeScanIterator("ns", "", ""))
	checkItrResults(t, itr3, []*statedb.VersionedKV{
		constructVersionedKV("ns", "key1", []byte("value1"), version.NewHeight(1, 1)),
		constructVersionedKV("ns", "key3", []byte("value3"), version.NewHeight(1, 1)),
		constructVersionedKV("ns", "key4", []byte("value4"), version.NewHeight(1, 1)),
		constructVersionedKV("ns", "key6", []byte("value6_new"), version.NewHeight(1, 1)),
		constructVersionedKV("ns", "key7", []byte("value7"), version.NewHeight(1, 1)),
	})
}

func checkItrResults(t *testing.T, itr statedb.ResultsIterator, expectedResults []*statedb.VersionedKV) {
	for i := 0; i < len(expectedResults); i++ {
		res, _ := itr.Next()
		testutil.AssertEquals(t, res, expectedResults[i])
	}
	lastRes, err := itr.Next()
	testutil.AssertNoError(t, err, "")
	testutil.AssertNil(t, lastRes)
}

func constructVersionedKV(ns string, key string, value []byte, version *version.Height) *statedb.VersionedKV {
	return &statedb.VersionedKV{
		CompositeKey:   statedb.CompositeKey{Namespace: ns, Key: key},
		VersionedValue: statedb.VersionedValue{Value: value, Version: version}}
}
