# zapsentry

Implementation of go.uber.org/zap/zapcore.Core for integration with Sentry.

## Example

```go
sentryClient, _ := sentry.NewClient(sentry.ClientOptions{
    Dsn:              "<SENTRY_DSN>",
    AttachStacktrace: true,
  })
log, _ = zap.NewProduction()

reqBodyCtxValName := "body"
log = log.
  WithOptions(
    zap.WrapCore(
      zapsentry.NewWrapper(
        sentryClient,
        zapsentry.WithRequest("request", &reqBodyCtxValName),
        zapsentry.WithSecretHeaders("Authorization-Token"),
      ),
    ),
  )
```