package correlation

import (
	"context"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi/v7/middlewares"
)

func ID(ctx context.Context) lager.Data {
	result := make(lager.Data)
	if cid, ok := ctx.Value(middlewares.CorrelationIDKey).(string); ok {
		result["correlation-id"] = cid
	}

	// When brokerapi supports the request ID:
	//if rid, ok := ctx.Value(middlewares.RequestIDKey).(string); ok {
	//	result["request-id"] = rid
	//}

	return result
}
