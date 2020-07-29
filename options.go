package zapsentry

import "strings"

type Option func(c *SentryCore)

func WithRequest(reqFieldName string, bodyCtxName *string) Option {
	return func(c *SentryCore) {
		c.requestFieldName = reqFieldName
		if bodyCtxName != nil {
			c.bodyCtxName = bodyCtxName
		}
	}
}

func WithSecretHeaders(secretHeaders ...string) Option {
	return func(c *SentryCore) {
		c.secretHeaders = make(map[string]struct{}, len(secretHeaders))

		for _, header := range secretHeaders {
			c.secretHeaders[strings.ToLower(header)] = struct{}{}
		}
	}
}
