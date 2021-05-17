// Copyright 2020 Pivotal Software, Inc.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//    http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package credstore

import (
	"fmt"
	"io"
	"os"
	"strings"

	"code.cloudfoundry.org/credhub-cli/credhub/permissions"
)

type credHubStoreMock struct{}

func (c *credHubStoreMock) Put(key string, credentials interface{}) (interface{}, error) {
	return nil, nil
}

func getFileName(key string) string {
	sub := string(key[14:])
	fileName := strings.ReplaceAll(sub, "/", "-")
	return fileName
}

func (c *credHubStoreMock) PutValue(key string, credentials interface{}) (interface{}, error) {
	credstorePath := os.Getenv("DEV_MODE_ONLY")
	if _, err := os.Stat(credstorePath); os.IsNotExist(err) {
		err = os.MkdirAll(credstorePath, 0777)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory for mock credstore: %w", err)
		}
	}
	file, err := os.Create(credstorePath + "/" + getFileName(key))
	if err != nil {
		return nil, fmt.Errorf("failed to create file for mock credstore: %w", err)
	}

	s := credentials.(string)
	b := []byte(s)

	_, err = file.Write(b)
	if err != nil {
		return nil, fmt.Errorf("failed to write to file for mock credstore: %w", err)
	}

	file.Close()

	return nil, nil
}

func (c *credHubStoreMock) GetValue(key string) (string, error) {
	credstorePath := os.Getenv("DEV_MODE_ONLY")
	file, err := os.Open(credstorePath + "/" + getFileName(key))
	if err != nil {
		return "", fmt.Errorf("failed to open file for mock credstore: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file for mock credstore: %w", err)
	}

	return string(content), nil
}

func (c *credHubStoreMock) Get(key string) (interface{}, error) {
	return nil, nil
}

func (c *credHubStoreMock) Delete(key string) error {
	return nil
}

func (c *credHubStoreMock) AddPermission(path string, actor string, ops []string) (*permissions.Permission, error) {
	return nil, nil
}

func (c *credHubStoreMock) DeletePermission(path string) error {
	return nil
}
