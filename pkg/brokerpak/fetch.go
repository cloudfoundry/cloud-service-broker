// Copyright 2018 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package brokerpak

import (
	"fmt"
	"os"
	"path/filepath"

	getter "github.com/hashicorp/go-getter"
)

// fetchArchive uses go-getter to download archives. By default go-getter
// decompresses archives, so this configuration prevents that.
func fetchArchive(src, dest string) error {
	return newFileGetterClient(src, dest).Get()
}

// fetchBrokerpak downloads;
// Relative paths are resolved relative to the executable.
func fetchBrokerpak(src, dest string) error {
	execWd := filepath.Dir(os.Args[0])
	execDir, err := filepath.Abs(execWd)
	if err != nil {
		return fmt.Errorf("couldn't turn dir %q into abs path: %v", execWd, err)
	}

	client := newFileGetterClient(src, dest)
	client.Pwd = execDir

	return client.Get()
}

func defaultGetters() map[string]getter.Getter {
	getters := map[string]getter.Getter{}
	for k, g := range getter.Getters {
		getters[k] = g
	}

	return getters
}

// newFileGetterClient creates a new client that will fetch a single file,
// with the default set of getters and will NOT automatically decompress it.
func newFileGetterClient(src, dest string) *getter.Client {
	return &getter.Client{
		Src: src,
		Dst: dest,

		Mode:          getter.ClientModeFile,
		Getters:       defaultGetters(),
		Decompressors: map[string]getter.Decompressor{},
	}
}
