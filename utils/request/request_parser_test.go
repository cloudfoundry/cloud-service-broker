package request_test

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"reflect"
	"testing"

	"code.cloudfoundry.org/brokerapi/v13/middlewares"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/request"
)

func TestDecodeOriginatingIdentityHeader(t *testing.T) {
	cases := []struct {
		name     string
		ctx      context.Context
		expected map[string]any
	}{
		{
			name: "good-header",
			ctx:  context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, "cloudfoundry eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="),
			expected: map[string]any{
				"platform": "cloudfoundry",
				"value": map[string]any{
					"user_id": "683ea748-3092-4ff4-b656-39cacc4d5360",
				},
			},
		},
		{
			name:     "no header",
			ctx:      context.Background(),
			expected: nil,
		},
		{
			name:     "wrong number of elements in header",
			ctx:      context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, "eyANCiAgInVzZXJfaWQiOiAiNjgzZWE3NDgtMzA5Mi00ZmY0LWI2NTYtMzljYWNjNGQ1MzYwIg0KfQ=="),
			expected: nil,
		},
		{
			name:     "non encoded value",
			ctx:      context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, `cloudfoundry { "user_id": "683ea748-3092-4ff4-b656-39cacc4d5360" }`),
			expected: nil,
		},
		{
			name:     "non json value",
			ctx:      context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, fmt.Sprintf("cloudfoundry %s", b64.StdEncoding.EncodeToString([]byte("not json")))),
			expected: nil,
		},
		{
			name:     "header is not a string",
			ctx:      context.WithValue(context.Background(), middlewares.OriginatingIdentityKey, 111),
			expected: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if actual := request.DecodeOriginatingIdentityHeader(tc.ctx); !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("DecodeOriginatingIdentityHeader() = %v, expected %v", actual, tc.expected)
			}
		})
	}
}
