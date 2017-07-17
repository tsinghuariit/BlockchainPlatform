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
	"context"
	"flag"
	"fmt"

	"google.golang.org/grpc"

	"github.com/hyperledger/fabric/orderer/localconfig"
	cb "github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
)

var conf *config.TopLevel

type broadcastClient struct {
	ab.AtomicBroadcast_BroadcastClient
}

func (bc *broadcastClient) broadcast(env *cb.Envelope) error {
	var err error
	var resp *ab.BroadcastResponse

	err = bc.Send(env)
	if err != nil {
		return err
	}

	resp, err = bc.Recv()
	if err != nil {
		return err
	}

	fmt.Println("Status:", resp)
	return nil
}

// cmdImpl holds the command and its arguments.
type cmdImpl struct {
	name string
	args argsImpl
}

// argsImpl holds all the possible arguments for all possible commands.
type argsImpl struct {
	consensusType  string
	creationPolicy string
	chainID        string
}

func init() {
	conf = config.Load()
}

func main() {
	cmd := new(cmdImpl)
	var srv string

	flag.StringVar(&srv, "server", fmt.Sprintf("%s:%d", conf.General.ListenAddress, conf.General.ListenPort), "The RPC server to connect to.")
	flag.StringVar(&cmd.name, "cmd", "newChain", "The action that this client is requesting via the config transaction.")
	flag.StringVar(&cmd.args.consensusType, "consensusType", conf.Genesis.OrdererType, "In case of a newChain command, the type of consensus the ordering service is running on.")
	flag.StringVar(&cmd.args.creationPolicy, "creationPolicy", "AcceptAllPolicy", "In case of a newChain command, the chain creation policy this request should be validated against.")
	flag.StringVar(&cmd.args.chainID, "chainID", "NewChainID", "In case of a newChain command, the chain ID to create.")
	flag.Parse()

	conn, err := grpc.Dial(srv, grpc.WithInsecure())
	defer func() {
		_ = conn.Close()
	}()
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}

	client, err := ab.NewAtomicBroadcastClient(conn).Broadcast(context.TODO())
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}

	bc := &broadcastClient{client}

	switch cmd.name {
	case "newChain":
		env := newChainRequest(cmd.args.consensusType, cmd.args.creationPolicy, cmd.args.chainID)
		fmt.Println("Requesting the creation of chain", cmd.args.chainID)
		fmt.Println(bc.broadcast(env))
	default:
		panic("Invalid command given")
	}
}
