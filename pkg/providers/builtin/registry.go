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

package builtin

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker"
)

// NOTE(josephlewis42) unless there are extenuating circumstances, as of 2019
// no new builtin providers should be added. Instead, providers should be
// added using downloadable brokerpaks.

// BuiltinBrokerRegistry creates a new registry with all the built-in brokers
// added to it.
func BuiltinBrokerRegistry() broker.BrokerRegistry {
	out := broker.BrokerRegistry{}
	RegisterBuiltinBrokers(out)
	return out
}

// RegisterBuiltinBrokers adds the built-in brokers to the given registry.
func RegisterBuiltinBrokers(registry broker.BrokerRegistry) {
}
