package zapsentry

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewWrapper(client *sentry.Client, options ...Option) func(zapcore.Core) zapcore.Core {
	return func(c zapcore.Core) zapcore.Core {
		return NewSentryCore(c, client, options...)
	}
}

type SentryCore struct {
	zapcore.Core

	client            *sentry.Client
	secretHeaders     map[string]struct{}
	requestFieldName  string
	bodyCtxName       *string
	minSentrySeverity zapcore.Level

	fields []zapcore.Field
}

func NewSentryCore(core zapcore.Core, client *sentry.Client, options ...Option) *SentryCore {
	c := &SentryCore{
		Core:              core,
		client:            client,
		minSentrySeverity: zapcore.DebugLevel,
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

func (c *SentryCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	fields = append(
		append(
			make([]zapcore.Field, 0, len(c.fields)+len(fields)),
			c.fields...,
		),
		fields...,
	)
	enc := zapcore.NewMapObjectEncoder()
	var reqVal interface{}
	{
		var j int
		for i := 0; i < len(fields); i++ {
			if fields[i].Key == c.requestFieldName {
				// if more 1, then use latest
				reqVal = fields[i].Interface

				continue
			}

			if j != i {
				fields[j] = fields[i]
			}

			j++

			fields[i].AddTo(enc)
		}

		if j < len(fields) {
			fields = fields[:j]
		}
	}

	var req *http.Request
	if reqVal != nil {
		var ok bool
		req, ok = reqVal.(*http.Request)
		if !ok {
			return fmt.Errorf("wrong type of request %v", reqVal)
		}

		fields = append(fields, zap.Any(c.requestFieldName, req))
	}

	if ent.Level >= c.minSentrySeverity {
		event := sentry.NewEvent()
		if req != nil {
			event.Request = createSentryRequest(req, c.secretHeaders, c.bodyCtxName)
		}
		event.Extra = enc.Fields
		event.Extra["stacktrace"] = ent.Stack

		event.Message = ent.Message
		event.Timestamp = ent.Time
		event.Level = sentrySeverity(ent.Level)

		if eventID := c.client.CaptureEvent(event, nil, sentry.CurrentHub().Scope()); eventID == nil {
			fields = append(fields, zap.String("sentry_error", "send event to sentry error"))
		}
	}

	if err := c.Core.Write(ent, fields); err != nil {
		return err
	}

	return nil
}

func (c *SentryCore) With(fields []zapcore.Field) zapcore.Core {
	return &SentryCore{
		Core:             c.Core.With(nil),
		client:           c.client,
		fields:           append(c.fields, fields...),
		bodyCtxName:      c.bodyCtxName,
		requestFieldName: c.requestFieldName,
		secretHeaders:    c.secretHeaders,
	}
}

func (c *SentryCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func createSentryRequest(req *http.Request, secretHeaders map[string]struct{}, bodyCtxName *string) *sentry.Request {
	sReq := &sentry.Request{
		URL:         req.URL.String(),
		Method:      req.Method,
		QueryString: req.URL.Query().Encode(),
		Headers:     make(map[string]string, len(req.Header)),
	}

	for name, values := range req.Header {
		for _, value := range values {
			if _, ok := secretHeaders[strings.ToLower(name)]; ok {
				continue
			}

			sReq.Headers[name] = value
		}
	}

	if bodyCtxName != nil {
		ctx := req.Context()
		reqBodyVal := ctx.Value(*bodyCtxName)

		// reqBody, ok := req.Context().Value(*bodyCtxName).(string)
		reqBodyBytes, ok := reqBodyVal.([]byte)
		if ok {
			sReq.Data = string(reqBodyBytes)
		}
	}

	return sReq
}

func sentrySeverity(lvl zapcore.Level) sentry.Level {
	switch lvl {
	case zapcore.DebugLevel:
		return sentry.LevelDebug
	case zapcore.InfoLevel:
		return sentry.LevelInfo
	case zapcore.WarnLevel:
		return sentry.LevelWarning
	case zapcore.ErrorLevel:
		return sentry.LevelError
	case zapcore.DPanicLevel:
		return sentry.LevelFatal
	case zapcore.PanicLevel:
		return sentry.LevelFatal
	case zapcore.FatalLevel:
		return sentry.LevelFatal
	default:
		// Unrecognized levels are fatal.
		return sentry.LevelFatal
	}
}
