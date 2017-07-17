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

package config

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestGoodConfig(t *testing.T) {
	config := Load()
	if config == nil {
		t.Fatalf("Could not load config")
	}
	t.Logf("%+v", config)
}

func TestBadConfig(t *testing.T) {
	config := viper.New()
	config.SetConfigName("orderer")
	config.AddConfigPath("../")

	err := config.ReadInConfig()
	if err != nil {
		t.Fatalf("Error reading %s plugin config: %s", Prefix, err)
	}

	var uconf struct{}

	err = ExactWithDateUnmarshal(config, &uconf)
	if err == nil {
		t.Fatalf("Should have failed to unmarshal")
	}
}

type testSlice struct {
	Inner struct {
		Slice []string
	}
}

func TestEnvSlice(t *testing.T) {
	envVar := "ORDERER_INNER_SLICE"
	envVal := "[a, b, c]"
	os.Setenv(envVar, envVal)
	defer os.Unsetenv(envVar)
	config := viper.New()
	config.SetEnvPrefix(Prefix)
	config.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)
	config.SetConfigType("yaml")

	data := "---\nInner:\n    Slice: [d,e,f]"

	err := config.ReadConfig(bytes.NewReader([]byte(data)))

	if err != nil {
		t.Fatalf("Error reading %s plugin config: %s", Prefix, err)
	}

	var uconf testSlice

	err = ExactWithDateUnmarshal(config, &uconf)
	if err != nil {
		t.Fatalf("Failed to unmarshal with: %s", err)
	}

	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(uconf.Inner.Slice, expected) {
		t.Fatalf("Did not get back the right slice, expeced: %v got %v", expected, uconf.Inner.Slice)
	}
}

type testByteSize struct {
	Inner struct {
		ByteSize uint32
	}
}

func TestByteSize(t *testing.T) {
	config := viper.New()
	config.SetConfigType("yaml")

	testCases := []struct {
		data     string
		expected uint32
	}{
		{"", 0},
		{"42", 42},
		{"42k", 42 * 1024},
		{"42kb", 42 * 1024},
		{"42K", 42 * 1024},
		{"42KB", 42 * 1024},
		{"42 K", 42 * 1024},
		{"42 KB", 42 * 1024},
		{"42m", 42 * 1024 * 1024},
		{"42mb", 42 * 1024 * 1024},
		{"42M", 42 * 1024 * 1024},
		{"42MB", 42 * 1024 * 1024},
		{"42 M", 42 * 1024 * 1024},
		{"42 MB", 42 * 1024 * 1024},
		{"3g", 3 * 1024 * 1024 * 1024},
		{"3gb", 3 * 1024 * 1024 * 1024},
		{"3G", 3 * 1024 * 1024 * 1024},
		{"3GB", 3 * 1024 * 1024 * 1024},
		{"3 G", 3 * 1024 * 1024 * 1024},
		{"3 GB", 3 * 1024 * 1024 * 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.data, func(t *testing.T) {
			data := fmt.Sprintf("---\nInner:\n    ByteSize: %s", tc.data)
			err := config.ReadConfig(bytes.NewReader([]byte(data)))
			if err != nil {
				t.Fatalf("Error reading config: %s", err)
			}
			var uconf testByteSize
			err = ExactWithDateUnmarshal(config, &uconf)
			if err != nil {
				t.Fatalf("Failed to unmarshal with: %s", err)
			}
			if uconf.Inner.ByteSize != tc.expected {
				t.Fatalf("Did not get back the right byte size, expeced: %v got %v", tc.expected, uconf.Inner.ByteSize)
			}
		})
	}
}

func TestByteSizeOverflow(t *testing.T) {
	config := viper.New()
	config.SetConfigType("yaml")

	data := "---\nInner:\n    ByteSize: 4GB"
	err := config.ReadConfig(bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatalf("Error reading config: %s", err)
	}
	var uconf testByteSize
	err = ExactWithDateUnmarshal(config, &uconf)
	if err == nil {
		t.Fatalf("Should have failed to unmarshal")
	}
}

// TestEnvInnerVar verifies that with the Unmarshal function that
// the environmental overrides still work on internal vars.  This was
// a bug in the original viper implementation that is worked around in
// the Load codepath for now
func TestEnvInnerVar(t *testing.T) {
	envVar1 := "ORDERER_GENERAL_LISTENPORT"
	envVal1 := uint16(80)
	envVar2 := "ORDERER_KAFKA_RETRY_PERIOD"
	envVal2 := "42s"
	os.Setenv(envVar1, fmt.Sprintf("%d", envVal1))
	os.Setenv(envVar2, envVal2)
	defer os.Unsetenv(envVar1)
	defer os.Unsetenv(envVar2)
	config := Load()

	if config == nil {
		t.Fatalf("Could not load config")
	}

	if config.General.ListenPort != envVal1 {
		t.Fatalf("Environmental override of inner config test 1 did not work")
	}
	v2, _ := time.ParseDuration(envVal2)
	if config.Kafka.Retry.Period != v2 {
		t.Fatalf("Environmental override of inner config test 2 did not work")
	}
}
