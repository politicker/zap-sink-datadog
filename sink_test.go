package sink

import (
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestSink(t *testing.T) {
	config := zap.NewProductionConfig()

	config.OutputPaths = []string{"dd://unused"}
	logger, _ := config.Build()
	defer logger.Sync()

	logger.Info("failed to fetch URL",
		// Structured context as strongly typed Field values.
		zap.String("url", "hi"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)
}
