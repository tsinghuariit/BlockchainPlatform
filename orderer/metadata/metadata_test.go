/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package metadata_test

import (
	"fmt"
	"runtime"
	"testing"

	common "github.com/hyperledger/fabric/common/metadata"
	"github.com/hyperledger/fabric/orderer/metadata"
	"github.com/stretchr/testify/assert"
)

func TestGetVersionInfo(t *testing.T) {
	expected := fmt.Sprintf("%s:\n Version: %s\n Go version: %s\n OS/Arch: %s",
		metadata.ProgramName, common.Version, runtime.Version(),
		fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH))
	assert.Equal(t, expected, metadata.GetVersionInfo())
}
