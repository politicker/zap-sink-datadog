package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"go.uber.org/zap"
)

type DataDogSink struct {
	ctx context.Context
	dd  *datadogV2.LogsApi
}

func (s DataDogSink) Write(p []byte) (int, error) {
	fmt.Println(string(p))

	body := []datadogV2.HTTPLogItem{
		{
			Ddsource: datadog.PtrString("nginx"),
			Ddtags:   datadog.PtrString("env:staging,version:5.1"),
			Hostname: datadog.PtrString("i-012345678"),
			Message:  string(p),
			Service:  datadog.PtrString("payment"),
		},
	}

	resp, r, err := s.dd.SubmitLog(s.ctx, body, *datadogV2.NewSubmitLogOptionalParameters().WithContentEncoding(datadogV2.CONTENTENCODING_DEFLATE))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LogsApi.SubmitLog`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	responseContent, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Fprintf(os.Stderr, "Response from `LogsApi.SubmitLog`:\n%s\n", responseContent)

	return 0, nil
}

func (s DataDogSink) Sync() error {
	return nil
}

func (s DataDogSink) Close() error {
	return nil
}

func New() *DataDogSink {
	ctx := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: os.Getenv("DD_API_KEY"),
			},
			"appKeyAuth": {
				Key: os.Getenv("DD_APP_KEY"),
			},
		},
	)

	ctx = context.WithValue(ctx,
		datadog.ContextServerVariables,
		map[string]string{
			"site": "us5.datadoghq.com",
		})

	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV2.NewLogsApi(apiClient)
	validate(ctx)

	return &DataDogSink{
		ctx: ctx,
		dd:  api,
	}
}

func validate(ctx context.Context) {
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewAuthenticationApi(apiClient)
	resp, r, err := api.Validate(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthenticationApi.Validate`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	responseContent, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Fprintf(os.Stdout, "Response from `AuthenticationApi.Validate`:\n%s\n", responseContent)
}

func init() {
	zap.RegisterSink("dd", func(u *url.URL) (zap.Sink, error) {
		return New(), nil
	})
}
