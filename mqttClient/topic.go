package mqttClient

import (
	"strings"
)

func getAvailabilityTopic(config Config) string {
	return replaceTemplate(config.AvailabilityTopic(), config)
}

func (c *Client) ReplaceTemplate(template string) string {
	return replaceTemplate(template, c.cfg)
}

func replaceTemplate(template string, cfg Config) (r string) {
	r = strings.Replace(template, "%Prefix%", cfg.TopicPrefix(), 1)
	r = strings.Replace(r, "%ClientId%", cfg.ClientId(), 1)
	return
}
