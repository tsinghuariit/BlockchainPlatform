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

package fileledger

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	ordererledger "github.com/hyperledger/fabric/orderer/ledger"
	cb "github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	"github.com/op/go-logging"

	"github.com/golang/protobuf/jsonpb"
)

var logger = logging.MustGetLogger("ordererledger/fileledger")
var closedChan chan struct{}

func init() {
	closedChan = make(chan struct{})
	close(closedChan)
}

const (
	blockFileFormatString      = "block_%020d.json"
	chainDirectoryFormatString = "chain_%s"
)

type cursor struct {
	fl          *fileLedger
	blockNumber uint64
}

type fileLedger struct {
	directory      string
	fqFormatString string
	height         uint64
	signal         chan struct{}
	lastHash       []byte
	marshaler      *jsonpb.Marshaler
}

type fileLedgerFactory struct {
	directory string
	ledgers   map[string]ordererledger.ReadWriter
	mutex     sync.Mutex
}

// New creates a new fileledger Factory and the ordering system chain specified by the systemGenesis block (if it does not already exist)
func New(directory string) ordererledger.Factory {

	logger.Debugf("Initializing fileLedger at '%s'", directory)
	if err := os.MkdirAll(directory, 0700); err != nil {
		logger.Fatalf("Could not create directory %s: %s", directory, err)
	}

	flf := &fileLedgerFactory{
		directory: directory,
		ledgers:   make(map[string]ordererledger.ReadWriter),
	}

	infos, err := ioutil.ReadDir(flf.directory)
	if err != nil {
		logger.Panicf("Error reading from directory %s while initializing fileledger: %s", flf.directory, err)
	}

	for _, info := range infos {
		if !info.IsDir() {
			continue
		}
		var chainID string
		_, err := fmt.Sscanf(info.Name(), chainDirectoryFormatString, &chainID)
		if err != nil {
			continue
		}
		fl, err := flf.GetOrCreate(chainID)
		if err != nil {
			logger.Warningf("Failed to initialize chain from %s:", err)
			continue
		}
		flf.ledgers[chainID] = fl
	}

	return flf
}

func (flf *fileLedgerFactory) ChainIDs() []string {
	flf.mutex.Lock()
	defer flf.mutex.Unlock()
	ids := make([]string, len(flf.ledgers))

	i := 0
	for key := range flf.ledgers {
		ids[i] = key
		i++
	}

	return ids
}

func (flf *fileLedgerFactory) GetOrCreate(chainID string) (ordererledger.ReadWriter, error) {
	flf.mutex.Lock()
	defer flf.mutex.Unlock()

	key := chainID

	l, ok := flf.ledgers[key]
	if ok {
		return l, nil
	}

	directory := fmt.Sprintf("%s/"+chainDirectoryFormatString, flf.directory, chainID)

	logger.Debugf("Initializing chain at '%s'", directory)

	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}

	ch := newChain(directory)
	flf.ledgers[key] = ch
	return ch, nil
}

// newChain creates a new chain backed by a file ledger
func newChain(directory string) ordererledger.ReadWriter {
	fl := &fileLedger{
		directory:      directory,
		fqFormatString: directory + "/" + blockFileFormatString,
		signal:         make(chan struct{}),
		marshaler:      &jsonpb.Marshaler{Indent: "  "},
	}
	fl.initializeBlockHeight()
	logger.Debugf("Initialized to block height %d with hash %x", fl.height-1, fl.lastHash)
	return fl
}

// initializeBlockHeight verifies all blocks exist between 0 and the block height, and populates the lastHash
func (fl *fileLedger) initializeBlockHeight() {
	infos, err := ioutil.ReadDir(fl.directory)
	if err != nil {
		panic(err)
	}
	nextNumber := uint64(0)
	for _, info := range infos {
		if info.IsDir() {
			continue
		}
		var number uint64
		_, err := fmt.Sscanf(info.Name(), blockFileFormatString, &number)
		if err != nil {
			continue
		}
		if number != nextNumber {
			panic(fmt.Errorf("Missing block %d in the chain", nextNumber))
		}
		nextNumber++
	}
	fl.height = nextNumber
	if fl.height == 0 {
		return
	}
	block, found := fl.readBlock(fl.height - 1)
	if !found {
		panic(fmt.Errorf("Block %d was in directory listing but error reading", fl.height-1))
	}
	if block == nil {
		panic(fmt.Errorf("Error reading block %d", fl.height-1))
	}
	fl.lastHash = block.Header.Hash()
}

// blockFilename returns the fully qualified path to where a block of a given number should be stored on disk
func (fl *fileLedger) blockFilename(number uint64) string {
	return fmt.Sprintf(fl.fqFormatString, number)
}

// writeBlock commits a block to disk
func (fl *fileLedger) writeBlock(block *cb.Block) {
	file, err := os.Create(fl.blockFilename(block.Header.Number))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = fl.marshaler.Marshal(file, block)
	logger.Debugf("Wrote block %d", block.Header.Number)
	if err != nil {
		panic(err)
	}

}

// readBlock returns the block or nil, and whether the block was found or not, (nil,true) generally indicates an irrecoverable problem
func (fl *fileLedger) readBlock(number uint64) (*cb.Block, bool) {
	file, err := os.Open(fl.blockFilename(number))
	if err == nil {
		defer file.Close()
		block := &cb.Block{}
		err = jsonpb.Unmarshal(file, block)
		if err != nil {
			return nil, true
		}
		logger.Debugf("Read block %d", block.Header.Number)
		return block, true
	}
	return nil, false
}

// Height returns the highest block number in the chain, plus one
func (fl *fileLedger) Height() uint64 {
	return fl.height
}

// Append appends a new block to the ledger
func (fl *fileLedger) Append(block *cb.Block) error {
	if block.Header.Number != fl.height {
		return fmt.Errorf("Block number should have been %d but was %d", fl.height, block.Header.Number)
	}

	if !bytes.Equal(block.Header.PreviousHash, fl.lastHash) {
		return fmt.Errorf("Block should have had previous hash of %x but was %x", fl.lastHash, block.Header.PreviousHash)
	}

	fl.writeBlock(block)
	fl.lastHash = block.Header.Hash()
	fl.height++
	close(fl.signal)
	fl.signal = make(chan struct{})
	return nil
}

// Iterator implements the ordererledger.Reader definition
func (fl *fileLedger) Iterator(startPosition *ab.SeekPosition) (ordererledger.Iterator, uint64) {
	switch start := startPosition.Type.(type) {
	case *ab.SeekPosition_Oldest:
		return &cursor{fl: fl, blockNumber: 0}, 0
	case *ab.SeekPosition_Newest:
		high := fl.height - 1
		return &cursor{fl: fl, blockNumber: high}, high
	case *ab.SeekPosition_Specified:
		if start.Specified.Number > fl.height {
			return &ordererledger.NotFoundErrorIterator{}, 0
		}
		return &cursor{fl: fl, blockNumber: start.Specified.Number}, start.Specified.Number
	}

	// This line should be unreachable, but the compiler requires it
	return &ordererledger.NotFoundErrorIterator{}, 0
}

// Next blocks until there is a new block available, or returns an error if the next block is no longer retrievable
func (cu *cursor) Next() (*cb.Block, cb.Status) {
	// This only loops once, as signal reading indicates the new block has been written
	for {
		block, found := cu.fl.readBlock(cu.blockNumber)
		if found {
			if block == nil {
				return nil, cb.Status_SERVICE_UNAVAILABLE
			}
			cu.blockNumber++
			return block, cb.Status_SUCCESS
		}
		<-cu.fl.signal
	}
}

// ReadyChan returns a channel that will close when Next is ready to be called without blocking
func (cu *cursor) ReadyChan() <-chan struct{} {
	signal := cu.fl.signal
	if _, err := os.Stat(cu.fl.blockFilename(cu.blockNumber)); os.IsNotExist(err) {
		return signal
	}
	return closedChan
}
