package sink

import (
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestSink(t *testing.T) {
	config := zap.NewProductionConfig()

	config.OutputPaths = []string{"dd://us5.datadoghq.com/test-service?source=test&hostname=local"}
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("crop service name",
		// Structured context as strongly typed Field values.
		zap.String("event", "event-name"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)
}
