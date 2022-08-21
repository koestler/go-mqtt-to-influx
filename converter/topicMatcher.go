package converter

import (
	"fmt"
	"regexp"
	"strings"
)

type TopicMatcherConfig interface {
	Topic() string
	Device() string
	DeviceIsDynamic() bool
}

type TopicMatcher interface {
	MatchDevice(messageTopic string) (device string, err error)
	GetSubscribeTopic() string
}

type topicMatcherStruct struct {
	cfg     TopicMatcherConfig
	matcher *regexp.Regexp
}

func CreateTopicMatcher(cfg TopicMatcherConfig) (TopicMatcher, error) {
	// create regexp to match against
	deviceExpr := regexp.QuoteMeta(cfg.Device())
	if cfg.DeviceIsDynamic() {
		deviceExpr = "(" + strings.ReplaceAll(deviceExpr, "\\+", "[^\\/]+") + ")"
	}

	// must not have anything before / after
	expr := "^" + strings.Replace(regexp.QuoteMeta(cfg.Topic()), "%Device%", deviceExpr, 1) + "$"

	matcher, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("cannot create topic machter: invalid regexp: %s", err)
	}

	return topicMatcherStruct{
		cfg:     cfg,
		matcher: matcher,
	}, nil
}

func (t topicMatcherStruct) MatchDevice(messageTopic string) (device string, err error) {
	matches := t.matcher.FindStringSubmatch(messageTopic)
	if matches == nil {
		err = fmt.Errorf("messageTopic='%s' does not match", messageTopic)
	} else {
		device = matches[1]
	}

	return
}

func (t topicMatcherStruct) GetSubscribeTopic() string {
	if !t.cfg.DeviceIsDynamic() {
		return t.cfg.Device()
	} else {
		return strings.Replace(t.cfg.Topic(), "%Device%", t.cfg.Device(), 1)
	}
}
