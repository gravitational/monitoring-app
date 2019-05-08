// Copyright 2015 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

// This file contains definitions of Alertmanager configuration objects
// copied from its source tree with some minor modifications:
//
// https://github.com/prometheus/alertmanager/blob/v0.17.0/config/config.go
// https://github.com/prometheus/alertmanager/blob/v0.17.0/config/notifiers.go
//
// In the original source code all objects have custom marshalers/unmarshalers
// which do not work well for our use-case because watcher has to be able to
// update Alertmanager configuration. For example, certain 'secret' fields
// (such as SMTP password) are not marshaled, or config fails unmarshaling
// if it fails validation so it's impossible for user/watcher to repair it.
//
// For those reasons, the config structs have been copied as-is but without
// custom marshalers and with 'secret' fields replaced with regular strings.

import (
	"net/url"
	"regexp"
	"time"

	"github.com/gravitational/trace"
	commoncfg "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

// Config is the top-level configuration for Alertmanager's config files.
type Config struct {
	Global       *GlobalConfig  `yaml:"global,omitempty" json:"global,omitempty"`
	Route        *Route         `yaml:"route,omitempty" json:"route,omitempty"`
	InhibitRules []*InhibitRule `yaml:"inhibit_rules,omitempty" json:"inhibit_rules,omitempty"`
	Receivers    []*Receiver    `yaml:"receivers,omitempty" json:"receivers,omitempty"`
	Templates    []string       `yaml:"templates" json:"templates"`
}

// Load parses the YAML input s into a Config.
func Load(s string) (*Config, error) {
	cfg := &Config{}
	err := yaml.UnmarshalStrict([]byte(s), cfg)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return cfg, nil
}

// String marshals config into YAML.
func (c Config) String() (string, error) {
	b, err := yaml.Marshal(c)
	if err != nil {
		return "", trace.Wrap(err)
	}
	return string(b), nil
}

// URL is a custom type that represents an HTTP or HTTPS URL and allows validation at configuration load time.
type URL struct {
	*url.URL
}

// GlobalConfig defines configuration parameters that are valid globally
// unless overwritten.
type GlobalConfig struct {
	// ResolveTimeout is the time after which an alert is declared resolved
	// if it has not been updated.
	ResolveTimeout model.Duration `yaml:"resolve_timeout" json:"resolve_timeout"`

	HTTPConfig *commoncfg.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	SMTPFrom         string `yaml:"smtp_from,omitempty" json:"smtp_from,omitempty"`
	SMTPHello        string `yaml:"smtp_hello,omitempty" json:"smtp_hello,omitempty"`
	SMTPSmarthost    string `yaml:"smtp_smarthost,omitempty" json:"smtp_smarthost,omitempty"`
	SMTPAuthUsername string `yaml:"smtp_auth_username,omitempty" json:"smtp_auth_username,omitempty"`
	SMTPAuthPassword string `yaml:"smtp_auth_password,omitempty" json:"smtp_auth_password,omitempty"`
	SMTPAuthSecret   string `yaml:"smtp_auth_secret,omitempty" json:"smtp_auth_secret,omitempty"`
	SMTPAuthIdentity string `yaml:"smtp_auth_identity,omitempty" json:"smtp_auth_identity,omitempty"`
	SMTPRequireTLS   bool   `yaml:"smtp_require_tls,omitempty" json:"smtp_require_tls,omitempty"`
	SlackAPIURL      *URL   `yaml:"slack_api_url,omitempty" json:"slack_api_url,omitempty"`
	PagerdutyURL     *URL   `yaml:"pagerduty_url,omitempty" json:"pagerduty_url,omitempty"`
	HipchatAPIURL    *URL   `yaml:"hipchat_api_url,omitempty" json:"hipchat_api_url,omitempty"`
	HipchatAuthToken string `yaml:"hipchat_auth_token,omitempty" json:"hipchat_auth_token,omitempty"`
	OpsGenieAPIURL   *URL   `yaml:"opsgenie_api_url,omitempty" json:"opsgenie_api_url,omitempty"`
	OpsGenieAPIKey   string `yaml:"opsgenie_api_key,omitempty" json:"opsgenie_api_key,omitempty"`
	WeChatAPIURL     *URL   `yaml:"wechat_api_url,omitempty" json:"wechat_api_url,omitempty"`
	WeChatAPISecret  string `yaml:"wechat_api_secret,omitempty" json:"wechat_api_secret,omitempty"`
	WeChatAPICorpID  string `yaml:"wechat_api_corp_id,omitempty" json:"wechat_api_corp_id,omitempty"`
	VictorOpsAPIURL  *URL   `yaml:"victorops_api_url,omitempty" json:"victorops_api_url,omitempty"`
	VictorOpsAPIKey  string `yaml:"victorops_api_key,omitempty" json:"victorops_api_key,omitempty"`
}

// A Route is a node that contains definitions of how to handle alerts.
type Route struct {
	Receiver string `yaml:"receiver,omitempty" json:"receiver,omitempty"`

	GroupByStr []string          `yaml:"group_by,omitempty" json:"group_by,omitempty"`
	GroupBy    []model.LabelName `yaml:"-" json:"-"`
	GroupByAll bool              `yaml:"-" json:"-"`

	Match    map[string]string `yaml:"match,omitempty" json:"match,omitempty"`
	MatchRE  map[string]Regexp `yaml:"match_re,omitempty" json:"match_re,omitempty"`
	Continue bool              `yaml:"continue,omitempty" json:"continue,omitempty"`
	Routes   []*Route          `yaml:"routes,omitempty" json:"routes,omitempty"`

	GroupWait      *model.Duration `yaml:"group_wait,omitempty" json:"group_wait,omitempty"`
	GroupInterval  *model.Duration `yaml:"group_interval,omitempty" json:"group_interval,omitempty"`
	RepeatInterval *model.Duration `yaml:"repeat_interval,omitempty" json:"repeat_interval,omitempty"`
}

// InhibitRule defines an inhibition rule that mutes alerts that match the
// target labels if an alert matching the source labels exists.
// Both alerts have to have a set of labels being equal.
type InhibitRule struct {
	// SourceMatch defines a set of labels that have to equal the given
	// value for source alerts.
	SourceMatch map[string]string `yaml:"source_match,omitempty" json:"source_match,omitempty"`
	// SourceMatchRE defines pairs like SourceMatch but does regular expression
	// matching.
	SourceMatchRE map[string]Regexp `yaml:"source_match_re,omitempty" json:"source_match_re,omitempty"`
	// TargetMatch defines a set of labels that have to equal the given
	// value for target alerts.
	TargetMatch map[string]string `yaml:"target_match,omitempty" json:"target_match,omitempty"`
	// TargetMatchRE defines pairs like TargetMatch but does regular expression
	// matching.
	TargetMatchRE map[string]Regexp `yaml:"target_match_re,omitempty" json:"target_match_re,omitempty"`
	// A set of labels that must be equal between the source and target alert
	// for them to be a match.
	Equal model.LabelNames `yaml:"equal,omitempty" json:"equal,omitempty"`
}

// Receiver configuration provides configuration on how to contact a receiver.
type Receiver struct {
	// A unique identifier for this receiver.
	Name string `yaml:"name" json:"name"`

	EmailConfigs     []*EmailConfig     `yaml:"email_configs,omitempty" json:"email_configs,omitempty"`
	PagerdutyConfigs []*PagerdutyConfig `yaml:"pagerduty_configs,omitempty" json:"pagerduty_configs,omitempty"`
	HipchatConfigs   []*HipchatConfig   `yaml:"hipchat_configs,omitempty" json:"hipchat_configs,omitempty"`
	SlackConfigs     []*SlackConfig     `yaml:"slack_configs,omitempty" json:"slack_configs,omitempty"`
	WebhookConfigs   []*WebhookConfig   `yaml:"webhook_configs,omitempty" json:"webhook_configs,omitempty"`
	OpsGenieConfigs  []*OpsGenieConfig  `yaml:"opsgenie_configs,omitempty" json:"opsgenie_configs,omitempty"`
	WechatConfigs    []*WechatConfig    `yaml:"wechat_configs,omitempty" json:"wechat_configs,omitempty"`
	PushoverConfigs  []*PushoverConfig  `yaml:"pushover_configs,omitempty" json:"pushover_configs,omitempty"`
	VictorOpsConfigs []*VictorOpsConfig `yaml:"victorops_configs,omitempty" json:"victorops_configs,omitempty"`
}

// Regexp encapsulates a regexp.Regexp and makes it YAML marshalable.
type Regexp struct {
	*regexp.Regexp
}

// NotifierConfig contains base options common across all notifier configurations.
type NotifierConfig struct {
	VSendResolved bool `yaml:"send_resolved" json:"send_resolved"`
}

// EmailConfig configures notifications via mail.
type EmailConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	// Email address to notify.
	To           string              `yaml:"to,omitempty" json:"to,omitempty"`
	From         string              `yaml:"from,omitempty" json:"from,omitempty"`
	Hello        string              `yaml:"hello,omitempty" json:"hello,omitempty"`
	Smarthost    string              `yaml:"smarthost,omitempty" json:"smarthost,omitempty"`
	AuthUsername string              `yaml:"auth_username,omitempty" json:"auth_username,omitempty"`
	AuthPassword string              `yaml:"auth_password,omitempty" json:"auth_password,omitempty"`
	AuthSecret   string              `yaml:"auth_secret,omitempty" json:"auth_secret,omitempty"`
	AuthIdentity string              `yaml:"auth_identity,omitempty" json:"auth_identity,omitempty"`
	Headers      map[string]string   `yaml:"headers,omitempty" json:"headers,omitempty"`
	HTML         string              `yaml:"html,omitempty" json:"html,omitempty"`
	Text         string              `yaml:"text,omitempty" json:"text,omitempty"`
	RequireTLS   *bool               `yaml:"require_tls,omitempty" json:"require_tls,omitempty"`
	TLSConfig    commoncfg.TLSConfig `yaml:"tls_config,omitempty" json:"tls_config,omitempty"`
}

// PagerdutyConfig configures notifications via PagerDuty.
type PagerdutyConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commoncfg.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	ServiceKey  string            `yaml:"service_key,omitempty" json:"service_key,omitempty"`
	RoutingKey  string            `yaml:"routing_key,omitempty" json:"routing_key,omitempty"`
	URL         *URL              `yaml:"url,omitempty" json:"url,omitempty"`
	Client      string            `yaml:"client,omitempty" json:"client,omitempty"`
	ClientURL   string            `yaml:"client_url,omitempty" json:"client_url,omitempty"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Details     map[string]string `yaml:"details,omitempty" json:"details,omitempty"`
	Images      []PagerdutyImage  `yaml:"images,omitempty" json:"images,omitempty"`
	Links       []PagerdutyLink   `yaml:"links,omitempty" json:"links,omitempty"`
	Severity    string            `yaml:"severity,omitempty" json:"severity,omitempty"`
	Class       string            `yaml:"class,omitempty" json:"class,omitempty"`
	Component   string            `yaml:"component,omitempty" json:"component,omitempty"`
	Group       string            `yaml:"group,omitempty" json:"group,omitempty"`
}

// PagerdutyLink is a link
type PagerdutyLink struct {
	HRef string `yaml:"href,omitempty" json:"href,omitempty"`
	Text string `yaml:"text,omitempty" json:"text,omitempty"`
}

// PagerdutyImage is an image
type PagerdutyImage struct {
	Src  string `yaml:"src,omitempty" json:"src,omitempty"`
	Alt  string `yaml:"alt,omitempty" json:"alt,omitempty"`
	Text string `yaml:"text,omitempty" json:"text,omitempty"`
}

// SlackAction configures a single Slack action that is sent with each notification.
// See https://api.slack.com/docs/message-attachments#action_fields and https://api.slack.com/docs/message-buttons
// for more information.
type SlackAction struct {
	Type         string                  `yaml:"type,omitempty"  json:"type,omitempty"`
	Text         string                  `yaml:"text,omitempty"  json:"text,omitempty"`
	URL          string                  `yaml:"url,omitempty"   json:"url,omitempty"`
	Style        string                  `yaml:"style,omitempty" json:"style,omitempty"`
	Name         string                  `yaml:"name,omitempty"  json:"name,omitempty"`
	Value        string                  `yaml:"value,omitempty"  json:"value,omitempty"`
	ConfirmField *SlackConfirmationField `yaml:"confirm,omitempty"  json:"confirm,omitempty"`
}

// SlackConfirmationField protect users from destructive actions or particularly distinguished decisions
// by asking them to confirm their button click one more time.
// See https://api.slack.com/docs/interactive-message-field-guide#confirmation_fields for more information.
type SlackConfirmationField struct {
	Text        string `yaml:"text,omitempty"  json:"text,omitempty"`
	Title       string `yaml:"title,omitempty"  json:"title,omitempty"`
	OkText      string `yaml:"ok_text,omitempty"  json:"ok_text,omitempty"`
	DismissText string `yaml:"dismiss_text,omitempty"  json:"dismiss_text,omitempty"`
}

// SlackField configures a single Slack field that is sent with each notification.
// Each field must contain a title, value, and optionally, a boolean value to indicate if the field
// is short enough to be displayed next to other fields designated as short.
// See https://api.slack.com/docs/message-attachments#fields for more information.
type SlackField struct {
	Title string `yaml:"title,omitempty" json:"title,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
	Short *bool  `yaml:"short,omitempty" json:"short,omitempty"`
}

// SlackConfig configures notifications via Slack.
type SlackConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commoncfg.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APIURL *URL `yaml:"api_url,omitempty" json:"api_url,omitempty"`

	// Slack channel override, (like #other-channel or @username).
	Channel  string `yaml:"channel,omitempty" json:"channel,omitempty"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Color    string `yaml:"color,omitempty" json:"color,omitempty"`

	Title       string         `yaml:"title,omitempty" json:"title,omitempty"`
	TitleLink   string         `yaml:"title_link,omitempty" json:"title_link,omitempty"`
	Pretext     string         `yaml:"pretext,omitempty" json:"pretext,omitempty"`
	Text        string         `yaml:"text,omitempty" json:"text,omitempty"`
	Fields      []*SlackField  `yaml:"fields,omitempty" json:"fields,omitempty"`
	ShortFields bool           `yaml:"short_fields,omitempty" json:"short_fields,omitempty"`
	Footer      string         `yaml:"footer,omitempty" json:"footer,omitempty"`
	Fallback    string         `yaml:"fallback,omitempty" json:"fallback,omitempty"`
	CallbackID  string         `yaml:"callback_id,omitempty" json:"callback_id,omitempty"`
	IconEmoji   string         `yaml:"icon_emoji,omitempty" json:"icon_emoji,omitempty"`
	IconURL     string         `yaml:"icon_url,omitempty" json:"icon_url,omitempty"`
	ImageURL    string         `yaml:"image_url,omitempty" json:"image_url,omitempty"`
	ThumbURL    string         `yaml:"thumb_url,omitempty" json:"thumb_url,omitempty"`
	LinkNames   bool           `yaml:"link_names,omitempty" json:"link_names,omitempty"`
	Actions     []*SlackAction `yaml:"actions,omitempty" json:"actions,omitempty"`
}

// HipchatConfig configures notifications via Hipchat.
type HipchatConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commoncfg.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APIURL        *URL   `yaml:"api_url,omitempty" json:"api_url,omitempty"`
	AuthToken     string `yaml:"auth_token,omitempty" json:"auth_token,omitempty"`
	RoomID        string `yaml:"room_id,omitempty" json:"room_id,omitempty"`
	From          string `yaml:"from,omitempty" json:"from,omitempty"`
	Notify        bool   `yaml:"notify,omitempty" json:"notify,omitempty"`
	Message       string `yaml:"message,omitempty" json:"message,omitempty"`
	MessageFormat string `yaml:"message_format,omitempty" json:"message_format,omitempty"`
	Color         string `yaml:"color,omitempty" json:"color,omitempty"`
}

// WebhookConfig configures notifications via a generic webhook.
type WebhookConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commoncfg.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	// URL to send POST request to.
	URL *URL `yaml:"url" json:"url"`
}

// WechatConfig configures notifications via Wechat.
type WechatConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commoncfg.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APISecret string `yaml:"api_secret,omitempty" json:"api_secret,omitempty"`
	CorpID    string `yaml:"corp_id,omitempty" json:"corp_id,omitempty"`
	Message   string `yaml:"message,omitempty" json:"message,omitempty"`
	APIURL    *URL   `yaml:"api_url,omitempty" json:"api_url,omitempty"`
	ToUser    string `yaml:"to_user,omitempty" json:"to_user,omitempty"`
	ToParty   string `yaml:"to_party,omitempty" json:"to_party,omitempty"`
	ToTag     string `yaml:"to_tag,omitempty" json:"to_tag,omitempty"`
	AgentID   string `yaml:"agent_id,omitempty" json:"agent_id,omitempty"`
}

// OpsGenieConfig configures notifications via OpsGenie.
type OpsGenieConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commoncfg.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APIKey      string            `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	APIURL      *URL              `yaml:"api_url,omitempty" json:"api_url,omitempty"`
	Message     string            `yaml:"message,omitempty" json:"message,omitempty"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Source      string            `yaml:"source,omitempty" json:"source,omitempty"`
	Details     map[string]string `yaml:"details,omitempty" json:"details,omitempty"`
	Teams       string            `yaml:"teams,omitempty" json:"teams,omitempty"`
	Tags        string            `yaml:"tags,omitempty" json:"tags,omitempty"`
	Note        string            `yaml:"note,omitempty" json:"note,omitempty"`
	Priority    string            `yaml:"priority,omitempty" json:"priority,omitempty"`
}

// VictorOpsConfig configures notifications via VictorOps.
type VictorOpsConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commoncfg.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	APIKey            string            `yaml:"api_key" json:"api_key"`
	APIURL            *URL              `yaml:"api_url" json:"api_url"`
	RoutingKey        string            `yaml:"routing_key" json:"routing_key"`
	MessageType       string            `yaml:"message_type" json:"message_type"`
	StateMessage      string            `yaml:"state_message" json:"state_message"`
	EntityDisplayName string            `yaml:"entity_display_name" json:"entity_display_name"`
	MonitoringTool    string            `yaml:"monitoring_tool" json:"monitoring_tool"`
	CustomFields      map[string]string `yaml:"custom_fields,omitempty" json:"custom_fields,omitempty"`
}

type PushoverConfig struct {
	NotifierConfig `yaml:",inline" json:",inline"`

	HTTPConfig *commoncfg.HTTPClientConfig `yaml:"http_config,omitempty" json:"http_config,omitempty"`

	UserKey  string        `yaml:"user_key,omitempty" json:"user_key,omitempty"`
	Token    string        `yaml:"token,omitempty" json:"token,omitempty"`
	Title    string        `yaml:"title,omitempty" json:"title,omitempty"`
	Message  string        `yaml:"message,omitempty" json:"message,omitempty"`
	URL      string        `yaml:"url,omitempty" json:"url,omitempty"`
	URLTitle string        `yaml:"url_title,omitempty" json:"url_title,omitempty"`
	Sound    string        `yaml:"sound,omitempty" json:"sound,omitempty"`
	Priority string        `yaml:"priority,omitempty" json:"priority,omitempty"`
	Retry    time.Duration `yaml:"retry,omitempty" json:"retry,omitempty"`
	Expire   time.Duration `yaml:"expire,omitempty" json:"expire,omitempty"`
	HTML     bool          `yaml:"html,omitempty" json:"html,omitempty"`
}
