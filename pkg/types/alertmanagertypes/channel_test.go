package alertmanagertypes

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigFromChannels(t *testing.T) {
	testCases := []struct {
		name              string
		channels          Channels
		expectedRoutes    []map[string]any
		expectedReceivers []map[string]any
	}{
		{
			name: "OneEmailChannel",
			channels: Channels{
				{
					Name: "email-receiver",
					Type: "email",
					Data: `{"name":"email-receiver","email_configs":[{"to":"test@example.com"}]}`,
				},
			},
			expectedRoutes:    []map[string]any{{"receiver": "email-receiver", "continue": true}},
			expectedReceivers: []map[string]any{{"name": "default-receiver"}, {"name": "email-receiver", "email_configs": []any{map[string]any{"send_resolved": false, "to": "test@example.com", "from": "alerts@example.com", "hello": "localhost", "smarthost": "smtp.example.com:587", "require_tls": true, "tls_config": map[string]any{"insecure_skip_verify": false}}}}},
		},
		{
			name: "OneSlackChannel",
			channels: Channels{
				{
					Name: "slack-receiver",
					Type: "slack",
					Data: `{"name":"slack-receiver","slack_configs":[{"channel":"#alerts","api_url":"https://slack.com/api/test","send_resolved":true}]}`,
				},
			},
			expectedRoutes:    []map[string]any{{"receiver": "slack-receiver", "continue": true}},
			expectedReceivers: []map[string]any{{"name": "default-receiver"}, {"name": "slack-receiver", "slack_configs": []any{map[string]any{"send_resolved": true, "http_config": map[string]any{"tls_config": map[string]any{"insecure_skip_verify": false}, "follow_redirects": true, "enable_http2": true, "proxy_url": nil}, "api_url": "https://slack.com/api/test", "channel": "#alerts"}}}},
		},
		{
			name: "OnePagerdutyChannel",
			channels: Channels{
				{
					Name: "pagerduty-receiver",
					Type: "pagerduty",
					Data: `{"name":"pagerduty-receiver","pagerduty_configs":[{"service_key":"test"}]}`,
				},
			},
			expectedRoutes:    []map[string]any{{"receiver": "pagerduty-receiver", "continue": true}},
			expectedReceivers: []map[string]any{{"name": "default-receiver"}, {"name": "pagerduty-receiver", "pagerduty_configs": []any{map[string]any{"send_resolved": false, "http_config": map[string]any{"tls_config": map[string]any{"insecure_skip_verify": false}, "follow_redirects": true, "enable_http2": true, "proxy_url": nil}, "service_key": "test", "url": "https://events.pagerduty.com/v2/enqueue"}}}},
		},
		{
			name: "OnePagerdutyAndOneSlackChannel",
			channels: Channels{
				{
					Name: "pagerduty-receiver",
					Type: "pagerduty",
					Data: `{"name":"pagerduty-receiver","pagerduty_configs":[{"service_key":"test"}]}`,
				},
				{
					Name: "slack-receiver",
					Type: "slack",
					Data: `{"name":"slack-receiver","slack_configs":[{"channel":"#alerts","api_url":"https://slack.com/api/test","send_resolved":true}]}`,
				},
			},
			expectedRoutes:    []map[string]any{{"receiver": "pagerduty-receiver", "continue": true}, {"receiver": "slack-receiver", "continue": true}},
			expectedReceivers: []map[string]any{{"name": "default-receiver"}, {"name": "pagerduty-receiver", "pagerduty_configs": []any{map[string]any{"send_resolved": false, "http_config": map[string]any{"tls_config": map[string]any{"insecure_skip_verify": false}, "follow_redirects": true, "enable_http2": true, "proxy_url": nil}, "service_key": "test", "url": "https://events.pagerduty.com/v2/enqueue"}}}, {"name": "slack-receiver", "slack_configs": []any{map[string]any{"send_resolved": true, "http_config": map[string]any{"tls_config": map[string]any{"insecure_skip_verify": false}, "follow_redirects": true, "enable_http2": true, "proxy_url": nil}, "api_url": "https://slack.com/api/test", "channel": "#alerts"}}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewConfigFromChannels(
				GlobalConfig{
					ResolveTimeout: model.Duration(5 * time.Minute),
					SMTPHello:      "localhost",
					SMTPFrom:       "alerts@example.com",
					SMTPSmarthost:  config.HostPort{Host: "smtp.example.com", Port: "587"},
					SMTPRequireTLS: true,
				},
				RouteConfig{
					GroupByStr:     []string{"alertname"},
					GroupInterval:  5 * time.Minute,
					GroupWait:      30 * time.Second,
					RepeatInterval: 4 * time.Hour,
				},
				tc.channels,
				"1",
			)
			assert.NoError(t, err)

			routes, err := json.Marshal(c.alertmanagerConfig.Route.Routes)
			assert.NoError(t, err)
			var actualRoutes []map[string]any
			err = json.Unmarshal(routes, &actualRoutes)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.expectedRoutes, actualRoutes)

			receivers, err := json.Marshal(c.alertmanagerConfig.Receivers)
			assert.NoError(t, err)
			var actualReceivers []map[string]any
			err = json.Unmarshal(receivers, &actualReceivers)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tc.expectedReceivers, actualReceivers)
		})
	}
}
