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

package chaincode

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/hyperledger/fabric/common/util"
	protcommon "github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/spf13/cobra"
)

var chaincodeUpgradeCmd *cobra.Command

// upgradeCmd returns the cobra command for Chaincode Upgrade
func upgradeCmd(cf *ChaincodeCmdFactory) *cobra.Command {
	chaincodeUpgradeCmd = &cobra.Command{
		Use:       "upgrade",
		Short:     fmt.Sprintf("Upgrade chaincode."),
		Long:      fmt.Sprintf(`Upgrade an existing chaincode with the specified one. The new chaincode will immediately replace the existing chaincode upon the transaction committed.`),
		ValidArgs: []string{"1"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return chaincodeUpgrade(cmd, args, cf)
		},
	}

	return chaincodeUpgradeCmd
}

//upgrade the command via Endorser
func upgrade(cmd *cobra.Command, cf *ChaincodeCmdFactory) (*protcommon.Envelope, error) {
	spec, err := getChaincodeSpecification(cmd)
	if err != nil {
		return nil, err
	}

	cds, err := getChaincodeBytes(spec)
	if err != nil {
		return nil, fmt.Errorf("Error getting chaincode code %s: %s", chainFuncName, err)
	}

	creator, err := cf.Signer.Serialize()
	if err != nil {
		return nil, fmt.Errorf("Error serializing identity for %s: %s\n", cf.Signer.GetIdentifier(), err)
	}

	uuid := util.GenerateUUID()

	prop, err := utils.CreateUpgradeProposalFromCDS(uuid, chainID, cds, creator)
	if err != nil {
		return nil, fmt.Errorf("Error creating proposal %s: %s\n", chainFuncName, err)
	}
	logger.Debugf("Get upgrade proposal for chaincode <%v>", spec.ChaincodeID)

	var signedProp *pb.SignedProposal
	signedProp, err = utils.GetSignedProposal(prop, cf.Signer)
	if err != nil {
		return nil, fmt.Errorf("Error creating signed proposal  %s: %s\n", chainFuncName, err)
	}

	proposalResponse, err := cf.EndorserClient.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return nil, fmt.Errorf("Error endorsing %s: %s\n", chainFuncName, err)
	}
	logger.Debugf("endorse upgrade proposal, get response <%v>", proposalResponse.Response)

	if proposalResponse != nil {
		// assemble a signed transaction (it's an Envelope message)
		env, err := utils.CreateSignedTx(prop, cf.Signer, proposalResponse)
		if err != nil {
			return nil, fmt.Errorf("Could not assemble transaction, err %s", err)
		}
		logger.Debug("Get Signed envelope")
		return env, nil
	}

	return nil, nil
}

// chaincodeUpgrade upgrades the chaincode. On success, the new chaincode
// version is printed to STDOUT
func chaincodeUpgrade(cmd *cobra.Command, args []string, cf *ChaincodeCmdFactory) error {
	var err error
	if cf == nil {
		cf, err = InitCmdFactory()
		if err != nil {
			return err
		}
	}
	defer cf.BroadcastClient.Close()

	env, err := upgrade(cmd, cf)
	if err != nil {
		return err
	}

	if env != nil {
		logger.Debug("Send signed envelope to orderer")
		err = cf.BroadcastClient.Send(env)
		return err
	}

	return nil
}
