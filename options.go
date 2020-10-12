package zapsentry

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

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

func WithLevel(level zapcore.Level) Option {
	return func(c *SentryCore) {
		c.minSentrySeverity = level
	}
}
