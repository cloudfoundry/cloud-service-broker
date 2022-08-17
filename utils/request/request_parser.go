// Package request decodes the originating identity header
package request

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"strings"

	"github.com/pivotal-cf/brokerapi/v8/middlewares"
)

func DecodeOriginatingIdentityHeader(ctx context.Context) map[string]any {
	var originatingIdentityMap map[string]any

	originatingIdentityHeader := ctx.Value(middlewares.OriginatingIdentityKey)
	if originatingIdentityHeader != nil {
		if headerAsString, ok := originatingIdentityHeader.(string); ok {
			platform, value := parseHeader(headerAsString)
			if value != "" {
				if valueMap := unmarshallBase64JSON(value); valueMap != nil {
					originatingIdentityMap = map[string]any{
						"platform": platform,
						"value":    valueMap,
					}
				}
			}
		}
	}

	return originatingIdentityMap
}

func unmarshallBase64JSON(input string) (result map[string]any) {
	value, err := b64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(value, &result); err != nil {
		return nil
	}
	return
}

func parseHeader(input string) (name, value string) {
	headerParts := strings.Split(strings.TrimSpace(input), " ")
	if len(headerParts) != 2 {
		return "", ""
	}
	return headerParts[0], headerParts[1]
}
