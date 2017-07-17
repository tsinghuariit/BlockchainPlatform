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

package ramledger

import (
	"testing"

	"github.com/hyperledger/fabric/orderer/common/bootstrap/provisional"
	"github.com/hyperledger/fabric/orderer/localconfig"
	cb "github.com/hyperledger/fabric/protos/common"

	logging "github.com/op/go-logging"
)

var genesisBlock *cb.Block

func init() {
	logging.SetLevel(logging.DEBUG, "")
	genesisBlock = provisional.New(config.Load()).GenesisBlock()
}

func NewTestChain(maxSize int) *ramLedger {
	rlf := New(maxSize)
	chain, err := rlf.GetOrCreate(provisional.TestChainID)
	if err != nil {
		panic(err)
	}
	chain.Append(genesisBlock)
	return chain.(*ramLedger)
}

// TestAppend ensures that appending blocks stores only the maxSize most recent blocks
// Note that 'only' is applicable because the genesis block will be discarded
func TestAppend(t *testing.T) {
	maxSize := 3
	rl := NewTestChain(maxSize)
	var blocks []*cb.Block
	for i := 0; i < 3; i++ {
		blocks = append(blocks, &cb.Block{Header: &cb.BlockHeader{Number: uint64(i + 1)}})
		rl.appendBlock(blocks[i])
	}
	item := rl.oldest
	for i := 0; i < 3; i++ {
		if item.block == nil {
			t.Fatalf("Block for item %d should not be nil", i)
		}
		if item.block.Header.Number != blocks[i].Header.Number {
			t.Errorf("Expected block %d to be %d but got %d", i, blocks[i].Header.Number, item.block.Header.Number)
		}
		if i != 2 && item.next == nil {
			t.Fatalf("Next item should not be nil")
		} else {
			item = item.next
		}
	}
}

// TestSignal checks if the signal channel closes when an item is appended
func TestSignal(t *testing.T) {
	maxSize := 3
	rl := NewTestChain(maxSize)
	item := rl.newest
	select {
	case <-item.signal:
		t.Fatalf("There is no successor, there should be no signal to continue")
	default:
	}
	rl.appendBlock(&cb.Block{Header: &cb.BlockHeader{Number: 1}})
	select {
	case <-item.signal:
	default:
		t.Fatalf("There is a successor, there should be a signal to continue")
	}
}

// TestTruncatingSafety is intended to simulate a reader who fetches a reference to the oldest list item
// which is then pushed off the history by appending greater than the history size (here, 10 appends with
// a maxSize of 3).  We let the go garbage collector ensure the references still exist
func TestTruncationSafety(t *testing.T) {
	maxSize := 3
	newBlocks := 10
	rl := NewTestChain(maxSize)
	item := rl.newest
	for i := 0; i < newBlocks; i++ {
		rl.appendBlock(&cb.Block{Header: &cb.BlockHeader{Number: uint64(i + 1)}})
	}
	count := 0
	for item.next != nil {
		item = item.next
		count++
	}

	if count != newBlocks {
		t.Fatalf("The iterator should have found %d new blocks but found %d", newBlocks, count)
	}
}
