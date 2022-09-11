package mqttClient

import (
	"strings"
)

func (c *ClientStruct) AvailabilityEnabled() bool {
	return len(c.cfg.AvailabilityTopic()) > 0
}

func (c *ClientStruct) GetAvailabilityTopic() string {
	return replaceTemplate(c.cfg.AvailabilityTopic(), c.cfg)
}

func (c *ClientStruct) ReplaceTemplate(template string) string {
	return replaceTemplate(template, c.cfg)
}

func replaceTemplate(template string, cfg Config) (r string) {
	r = strings.Replace(template, "%Prefix%", cfg.TopicPrefix(), 1)
	r = strings.Replace(r, "%ClientId%", cfg.ClientId(), 1)
	return
}
