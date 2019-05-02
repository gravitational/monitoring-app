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

	"github.com/gravitational/trace"
	"github.com/prometheus/alertmanager/config"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Resources interface {
	UpdateSMTPConfig(SMTPConfig) error
	UpdateAlertTarget(AlertTarget) error
	DeleteAlertTarget() error
	CreateAlert(Alert) error
	DeleteAlert(name string) error
}

type Client struct {
	Config
}

// New returns a new resources manager client.
func New(conf Config) (*Client, error) {
	err := conf.CheckAndSetDefaults()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &Client{Config: conf}, nil
}

type Config struct {
	Kubernetes *kubernetes.Clientset
	logrus.FieldLogger
}

func (c *Config) CheckAndSetDefaults() error {
	if c.Kubernetes == nil {
		return trace.BadParameter("kubernetes client is not set")
	}
	if c.FieldLogger == nil {
		c.FieldLogger = logrus.WithField(trace.Component, "resources")
	}
	return nil
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

func (c *Client) CreateAlert(Alert) error {
	return nil
}

func (c *Client) DeleteAlert(name string) error {
	return nil
}

func (c *Client) getAlertmanagerConfig() (*config.Config, error) {
	secret, err := c.Kubernetes.CoreV1().Secrets("monitoring").Get("alertmanager-main", metav1.GetOptions{})
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
	secret, err := c.Kubernetes.CoreV1().Secrets("monitoring").Get("alertmanager-main", metav1.GetOptions{})
	if err != nil {
		return trace.Wrap(err)
	}
	secret.StringData["alertmanager.yaml"] = conf.String()
	_, err = c.Kubernetes.CoreV1().Secrets("monitoring").Update(secret)
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
