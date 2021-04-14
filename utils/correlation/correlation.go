package correlation

import (
	"context"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi/v8/middlewares"
)

func ID(ctx context.Context) lager.Data {
	result := make(lager.Data)
	if cid, ok := ctx.Value(middlewares.CorrelationIDKey).(string); ok {
		result["correlation-id"] = cid
	}

	if rid, ok := ctx.Value(middlewares.RequestIdentityKey).(string); ok {
		result["request-id"] = rid
	}

	return result
}
