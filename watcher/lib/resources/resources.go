/*
Copyright 2019 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"fmt"
	"time"

	v1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monitoring "github.com/coreos/prometheus-operator/pkg/client/versioned"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/gravitational/rigging"
	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Resources provides an interface for managing monitoring resources.
type Resources interface {
	// UpsertSMTPConfig creates or updates cluster SMTP configuration.
	UpsertSMTPConfig(SMTPConfig) error
	// DeleteSMTPConfig resets cluster SMTP configuration.
	DeleteSMTPConfig() error
	// UpsertAlertTarget creates or updates recipient of monitoring alerts.
	UpsertAlertTarget(AlertTarget) error
	// DeleteAlertTarget resets monitoring alerts recipient.
	DeleteAlertTarget() error
	// UpsertAlert creates a new or updates an existing monitoring alert.
	UpsertAlert(Alert) error
	// DeleteAlert deletes specified monitoring alert.
	DeleteAlert(name string) error
}

// Client is Prometheus-based monitoring resource manager.
//
// Implements Resources.
type Client struct {
	// Secrets is the Kubernetes Secrets client.
	Secrets corev1.SecretInterface
	// Rules is the Kubernetes PrometheusRules CRD client.
	Rules monitoringv1.PrometheusRuleInterface
	// Namespace is the monitoring namespace.
	Namespace string
	// FieldLogger provides logging facilities.
	logrus.FieldLogger
}

// ClientConfig is the client configuration.
type ClientConfig struct {
	// KubernetesClient is the Kubernetes API client.
	KubernetesClient *kubernetes.Clientset
	// MonitoringClient is the Kubernetes Prometheus CRD resources API client.
	MonitoringClient *monitoring.Clientset
	// Namespace is the monitoring namespace.
	Namespace string
}

// CheckAndSetDefaults validates client configuration and sets defaults.
func (c *ClientConfig) CheckAndSetDefaults() error {
	var errors []error
	if c.KubernetesClient == nil {
		errors = append(errors, trace.BadParameter("missing kubernetes client"))
	}
	if c.MonitoringClient == nil {
		errors = append(errors, trace.BadParameter("missing monitoring client"))
	}
	if c.Namespace == "" {
		errors = append(errors, trace.BadParameter("missing namespace"))
	}
	return trace.NewAggregate(errors...)
}

// New returns a new resources manager client.
func New(conf ClientConfig) (*Client, error) {
	err := conf.CheckAndSetDefaults()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &Client{
		Secrets:     conf.KubernetesClient.CoreV1().Secrets(conf.Namespace),
		Rules:       conf.MonitoringClient.MonitoringV1().PrometheusRules(conf.Namespace),
		Namespace:   conf.Namespace,
		FieldLogger: logrus.WithField(trace.Component, "resources"),
	}, nil
}

// SMTPConfig represents cluster SMTP configuration.
type SMTPConfig struct {
	// Host is the SMTP host.
	Host string
	// Port is the SMTP port.
	Port int
	// Username is the SMTP user name.
	Username string
	// Password is the SMTP user password.
	Password string
}

// String returns the config's string representation.
func (c SMTPConfig) String() string {
	return fmt.Sprintf("SMTP(Host=%v,Port=%v,Username=%v)", c.Host, c.Port, c.Username)
}

// AlertTarget represents a recipient of monitoring alerts.
type AlertTarget struct {
	// Email is the recipient email address.
	Email string
}

// String returns the target's string representation.
func (t AlertTarget) String() string {
	return fmt.Sprintf("AlertTarget(Email=%v)", t.Email)
}

// Alert represents a monitoring alert.
type Alert struct {
	// CRDName is the name of PrometheusRule custom resource.
	CRDName string
	// AlertName is the alerting rule name.
	AlertName string
	// GroupName is group name the alert belongs to.
	GroupName string
	// Formula is the alert expression.
	Formula string
	// Delay is an optional delay before alert should be triggerred.
	Delay time.Duration
	// Labels is the labels that get attached to the alert.
	Labels map[string]string
	// Annotations are used to attach longer information to the alert.
	Annotations map[string]string
}

// String returns the alert's string representation.
func (a Alert) String() string {
	return fmt.Sprintf("Alert(CRDName=%v,AlertName=%v,GroupName=%v,Formula=%v,Delay=%v,Labels=%v)",
		a.CRDName, a.AlertName, a.GroupName, a.Formula, a.Delay, a.Labels)
}

// UpsertSMTPConfig updates cluster SMTP configuration.
func (c *Client) UpsertSMTPConfig(smtpConf SMTPConfig) error {
	c.Infof("Updating SMTP configuration: %s.", smtpConf)
	conf, err := c.getAlertmanagerConfig()
	if err != nil {
		return trace.Wrap(err)
	}
	err = updateSMTPConfig(conf, fmt.Sprintf("%v:%v", smtpConf.Host, smtpConf.Port),
		smtpConf.Username, smtpConf.Password)
	if err != nil {
		return trace.Wrap(err)
	}
	err = c.updateAlertmanagerConfig(conf)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// DeleteSMTPConfig resets cluster SMTP configuration.
func (c *Client) DeleteSMTPConfig() error {
	c.Info("Deleting SMTP configuration.")
	conf, err := c.getAlertmanagerConfig()
	if err != nil {
		return trace.Wrap(err)
	}
	err = deleteSMTPConfig(conf)
	if err != nil {
		return trace.Wrap(err)
	}
	err = c.updateAlertmanagerConfig(conf)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// UpsertAlertTarget updates recipient of monitoring alerts.
func (c *Client) UpsertAlertTarget(alertTarget AlertTarget) error {
	c.Infof("Updating alert target: %s.", alertTarget)
	conf, err := c.getAlertmanagerConfig()
	if err != nil {
		return trace.Wrap(err)
	}
	// We expect "default" receiver to be present.
	defaultReceiver, err := getDefaultReceiver(conf)
	if err != nil {
		return trace.Wrap(err)
	}
	// Update its email config.
	defaultReceiver.EmailConfigs = []*EmailConfig{
		{To: alertTarget.Email},
	}
	err = c.updateAlertmanagerConfig(conf)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// DeleteAlertTarget resets monitoring alerts recipient.
func (c *Client) DeleteAlertTarget() error {
	c.Info("Deleting alert target.")
	conf, err := c.getAlertmanagerConfig()
	if err != nil {
		return trace.Wrap(err)
	}
	// We expect "default" receiver to be present.
	defaultReceiver, err := getDefaultReceiver(conf)
	if err != nil {
		return trace.Wrap(err)
	}
	// Reset its email config.
	defaultReceiver.EmailConfigs = nil
	err = c.updateAlertmanagerConfig(conf)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// UpsertAlert creates a new or updates an existing monitoring alert.
func (c *Client) UpsertAlert(alert Alert) error {
	c.Infof("Creating alert: %s.", alert)
	_, err := c.Rules.Create(c.newPrometheusRule(alert))
	if err == nil {
		return nil
	}
	err = rigging.ConvertError(err)
	if !trace.IsAlreadyExists(err) {
		return trace.Wrap(err)
	}
	// Updating PrometheusRule requires resourceVersion to be set on the
	// CRD object so retrieve it first and update appropriate fields.
	rule, err := c.Rules.Get(alert.CRDName, metav1.GetOptions{})
	if err != nil {
		return trace.Wrap(rigging.ConvertError(err))
	}
	c.updatePrometheusRule(rule, alert)
	_, err = c.Rules.Update(rule)
	if err != nil {
		return trace.Wrap(rigging.ConvertError(err))
	}
	return nil
}

// DeleteAlert deletes specified monitoring alert.
func (c *Client) DeleteAlert(name string) error {
	c.Infof("Deleting alert: %v.", name)
	err := c.Rules.Delete(name, nil)
	if err != nil {
		return trace.Wrap(rigging.ConvertError(err))
	}
	return nil
}

// getAlertmanagerConfig returns Alertmanager configuration.
func (c *Client) getAlertmanagerConfig() (*Config, error) {
	secret, err := c.Secrets.Get(alertmanagerSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, trace.Wrap(rigging.ConvertError(err))
	}
	confBytes, ok := secret.Data[alertmanagerConfigFilename]
	if !ok {
		return nil, trace.NotFound("no alert manager config found")
	}
	conf, err := Load(string(confBytes))
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return conf, nil
}

// updateAlertmanagerConfig updates Alertmanager configuration.
func (c *Client) updateAlertmanagerConfig(conf *Config) error {
	c.Debugf("Updating alertmanager configuration file: %#v.", conf)
	secret, err := c.Secrets.Get(alertmanagerSecretName, metav1.GetOptions{})
	if err != nil {
		return trace.Wrap(rigging.ConvertError(err))
	}
	confString, err := conf.String()
	if err != nil {
		return trace.Wrap(err)
	}
	secret.StringData = map[string]string{
		alertmanagerConfigFilename: confString,
	}
	_, err = c.Secrets.Update(secret)
	if err != nil {
		return trace.Wrap(rigging.ConvertError(err))
	}
	return nil
}

// getDefaultReceiver returns receiver with the name "default" from the config.
func getDefaultReceiver(conf *Config) (*Receiver, error) {
	for _, r := range conf.Receivers {
		if r.Name == "default" {
			return r, nil
		}
	}
	return nil, trace.NotFound("no default receiver")
}

// updateSMTPConfig updates SMTP configuration in the provided config.
func updateSMTPConfig(conf *Config, addr, user, pass string) error {
	conf.Global.SMTPSmarthost = addr
	conf.Global.SMTPAuthUsername = user
	conf.Global.SMTPAuthPassword = pass
	// Update SMTP config on the default receiver too.
	defaultReceiver, err := getDefaultReceiver(conf)
	if err != nil {
		return trace.Wrap(err)
	}
	for _, emailConfig := range defaultReceiver.EmailConfigs {
		emailConfig.Smarthost = addr
		emailConfig.AuthUsername = user
		emailConfig.AuthPassword = pass
	}
	return nil
}

// deleteSMTPConfig resets SMTP configuration in the provided config.
func deleteSMTPConfig(conf *Config) error {
	return updateSMTPConfig(conf, "", "", "")
}

// updatePrometheusRule updates the provided PrometheusRule spec based on
// the new alert data.
func (c *Client) updatePrometheusRule(rule *v1.PrometheusRule, alert Alert) {
	rule.Spec = c.newPrometheusRule(alert).Spec
}

// newPrometheusRule returns PrometheusRule CRD object for the provided alert.
func (c *Client) newPrometheusRule(alert Alert) *v1.PrometheusRule {
	groupName := alert.GroupName
	if groupName == "" {
		groupName = fmt.Sprintf("%v.rules", alert.CRDName)
	}
	alertName := alert.AlertName
	if alertName == "" {
		alertName = alert.CRDName
	}
	rule := v1.Rule{
		Alert:       alertName,
		Expr:        intstr.FromString(alert.Formula),
		Labels:      alert.Labels,
		Annotations: alert.Annotations,
	}
	return &v1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1.PrometheusRuleKind,
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      alert.CRDName,
			Namespace: c.Namespace,
			Labels:    prometheusRuleLabels,
		},
		Spec: v1.PrometheusRuleSpec{
			Groups: []v1.RuleGroup{{
				Name:  groupName,
				Rules: []v1.Rule{rule},
			}},
		},
	}
}

// prometheusRuleLabels is the labels that PrometheusRule CRD should be marked
// with in order to be recognized by Prometheus operator controller.
var prometheusRuleLabels = map[string]string{
	"prometheus": "k8s",
	"role":       "alert-rules",
}

// alertmanagerConfigFilename is the name of Alertmanager configuration file.
var alertmanagerConfigFilename = "alertmanager.yaml"

// alertmanagerSecretName is the name of the secret with Alertmanager configuration.
var alertmanagerSecretName = "alertmanager-main"
