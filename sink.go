package sink

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"go.uber.org/zap"
)

type DataDogSink struct {
	ctx      context.Context
	dd       *datadogV2.LogsApi
	service  string
	hostname string
	tags     string
	source   string
}

func (s DataDogSink) Write(p []byte) (int, error) {
	fmt.Println(string(p))

	body := []datadogV2.HTTPLogItem{
		{
			Ddsource: datadog.PtrString(s.source),
			Ddtags:   datadog.PtrString(s.tags),
			Hostname: datadog.PtrString(s.hostname),
			Message:  string(p),
			Service:  datadog.PtrString(s.service),
		},
	}

	_, r, err := s.dd.SubmitLog(s.ctx, body, *datadogV2.NewSubmitLogOptionalParameters().WithContentEncoding(datadogV2.CONTENTENCODING_DEFLATE))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LogsApi.SubmitLog`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	return 0, nil
}

func (s DataDogSink) Sync() error {
	return nil
}

func (s DataDogSink) Close() error {
	return nil
}

func New(site string, service string, hostname string, tags string, source string) (*DataDogSink, error) {
	apiKey := os.Getenv("DD_API_KEY")
	if apiKey == "" {
		return nil, errors.New("missing env DD_API_KEY")
	}

	// configure credentials for datadog api
	ctx := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: apiKey,
			},
		},
	)

	config := datadog.NewConfiguration()
	_, err := config.Servers.URL(0, map[string]string{"site": site})
	if err != nil {
		return nil, fmt.Errorf("site is invalid: %v", err)
	}

	// configure datadog api endpoint (site)
	ctx = context.WithValue(ctx,
		datadog.ContextServerVariables,
		map[string]string{
			"site": site,
		},
	)

	return &DataDogSink{
		ctx:      ctx,
		dd:       datadogV2.NewLogsApi(datadog.NewAPIClient(config)),
		service:  service,
		hostname: hostname,
		tags:     tags,
		source:   source,
	}, nil
}

func init() {
	err := zap.RegisterSink("dd", func(u *url.URL) (zap.Sink, error) {
		site := u.Host
		service := u.Path
		hostname := u.Query().Get("hostname")
		tags := u.Query().Get("tags")
		source := u.Query().Get("source")

		if source == "" {
			source = "zap-logger"
		}

		if service == "" {
			return nil, errors.New("missing service name")
		} else {
			service = service[1:]
		}

		if service == "" {
			return nil, errors.New("missing service name")
		}

		sink, err := New(site, service, hostname, tags, source)
		if err != nil {
			return nil, fmt.Errorf("init failed: %v", err)
		}

		return sink, nil
	})

	if err != nil {
		log.Fatal(err)
	}
}
