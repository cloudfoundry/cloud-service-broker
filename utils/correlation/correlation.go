// Package correlation reads correlation IDs from the context for logging
package correlation

import (
	"context"

	"code.cloudfoundry.org/brokerapi/v13/middlewares"
	"code.cloudfoundry.org/lager/v3"
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
