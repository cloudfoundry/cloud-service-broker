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

package models

import (
	"crypto/rand"
	"crypto/sha256"
	"io"
)

var Key [32]byte

// NewKey creates new encryption key
// TODO: take the key from env vars
// TODO: Create accessor for Key
// TODO: Add logging and error handling
func NewKey() {
	dbKey := make([]byte, 32)
	io.ReadFull(rand.Reader, dbKey)
	Key = sha256.Sum256(dbKey)
}
