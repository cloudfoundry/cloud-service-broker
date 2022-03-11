// Copyright 2019 the Service Broker Project Authors.
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

package base

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

// MergedInstanceCredsMixin adds the BuildInstanceCredentials function that
// merges the OtherDetails of the bind and instance records.
type MergedInstanceCredsMixin struct{}

// BuildInstanceCredentials combines the bind credentials with the connection
// information in the instance details to get a full set of connection details.
func (b *MergedInstanceCredsMixin) BuildInstanceCredentials(ctx context.Context, credentials map[string]interface{}, outputs storage.JSONObject) (*domain.Binding, error) {
	vc, err := varcontext.Builder().
		MergeMap(outputs).
		MergeMap(credentials).
		Build()
	if err != nil {
		return nil, err
	}

	return &domain.Binding{Credentials: vc.ToMap()}, nil
}
