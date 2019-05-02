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

	v1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/gravitational/trace"
	"github.com/prometheus/alertmanager/config"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Resources interface {
	UpdateSMTPConfig(SMTPConfig) error
	UpdateAlertTarget(AlertTarget) error
	DeleteAlertTarget() error
	CreateAlert(Alert) error
	DeleteAlert(name string) error
}

type Client struct {
	Secrets   corev1.SecretInterface
	Rules     monitoringv1.PrometheusRuleInterface
	Namespace string
	logrus.FieldLogger
}

type Config struct {
	Client    kubernetes.Interface
	Namespace string
}

func (c *Config) CheckAndSetDefaults() error {
	if c.Client == nil {
		return trace.BadParameter("missing kubernetes client")
	}
	if c.Namespace == "" {
		return trace.BadParameter("missing namespace")
	}
	return nil
}

// New returns a new resources manager client.
func New(conf Config) (*Client, error) {
	err := conf.CheckAndSetDefaults()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	restClient := conf.Client.Discovery().RESTClient()
	return &Client{
		Secrets:     conf.Client.CoreV1().Secrets(conf.Namespace),
		Rules:       monitoringv1.New(restClient).PrometheusRules(conf.Namespace),
		Namespace:   conf.Namespace,
		FieldLogger: logrus.WithField(trace.Component, "resources"),
	}, nil
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type AlertTarget struct {
	Email string
}

type Alert struct {
	Name    string
	Formula string
}

func (c *Client) UpdateSMTPConfig(smtpConf SMTPConfig) error {
	c.Infof("Updating SMTP configuration: %#v.", smtpConf)
	conf, err := c.getAlertmanagerConfig()
	if err != nil {
		return trace.Wrap(err)
	}
	conf.Global.SMTPSmarthost = fmt.Sprintf("%v:%v", smtpConf.Host, smtpConf.Port)
	conf.Global.SMTPAuthUsername = smtpConf.Username
	conf.Global.SMTPAuthPassword = config.Secret(smtpConf.Password)
	err = c.updateAlertmanagerConfig(conf)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func (c *Client) UpdateAlertTarget(alertTarget AlertTarget) error {
	c.Infof("Updating alert target: %#v.", alertTarget)
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
	defaultReceiver.EmailConfigs = []*config.EmailConfig{
		{To: alertTarget.Email},
	}
	err = c.updateAlertmanagerConfig(conf)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

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
	defaultReceiver.EmailConfigs = []*config.EmailConfig{}
	err = c.updateAlertmanagerConfig(conf)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func (c *Client) CreateAlert(alert Alert) error {
	c.Infof("Creating alert: %#v.", alert)
	rule := &v1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1.PrometheusRuleKind,
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      alert.Name,
			Namespace: c.Namespace,
		},
		Spec: v1.PrometheusRuleSpec{
			Groups: []v1.RuleGroup{{
				Name: fmt.Sprintf("%v.rules", alert.Name),
				Rules: []v1.Rule{{
					Alert: alert.Name,
					Expr:  intstr.FromString(alert.Formula),
				}}},
			},
		},
	}
	_, err := c.Rules.Create(rule)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func (c *Client) DeleteAlert(name string) error {
	c.Infof("Deleting alert: %v.", name)
	err := c.Rules.Delete(name, nil)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func (c *Client) getAlertmanagerConfig() (*config.Config, error) {
	secret, err := c.Secrets.Get("alertmanager-main", metav1.GetOptions{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	confBytes, ok := secret.Data["alertmanager.yaml"]
	if !ok {
		return nil, trace.NotFound("no alert manager config found")
	}
	conf, err := config.Load(string(confBytes))
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return conf, nil
}

func (c *Client) updateAlertmanagerConfig(conf *config.Config) error {
	c.Debugf("Updating alertmanager configuration file: %#v.", conf)
	secret, err := c.Secrets.Get("alertmanager-main", metav1.GetOptions{})
	if err != nil {
		return trace.Wrap(err)
	}
	secret.StringData["alertmanager.yaml"] = conf.String()
	_, err = c.Secrets.Update(secret)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// getDefaultReceiver returns receiver with the name "default" from the config.
func getDefaultReceiver(conf *config.Config) (*config.Receiver, error) {
	for _, r := range conf.Receivers {
		if r.Name == "default" {
			return r, nil
		}
	}
	return nil, trace.NotFound("no default receiver")
}
