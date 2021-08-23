module github.com/gravitational/monitoring-app/watcher

go 1.15

require (
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/gosimple/slug v1.1.1
	github.com/gravitational/rigging v0.0.0-20210822131352-bc158e784924
	github.com/gravitational/roundtrip v1.0.0
	github.com/gravitational/trace v1.1.6
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.50.0
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.50.0
	github.com/prometheus/common v0.10.0
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be // indirect
	github.com/sirupsen/logrus v1.6.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v12.0.0+incompatible
)

replace (
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring => github.com/gravitational/prometheus-operator/pkg/apis/monitoring v0.43.1-0.20210818162409-2906a7bf1935
	github.com/prometheus-operator/prometheus-operator/pkg/client => github.com/gravitational/prometheus-operator/pkg/client v0.0.0-20210818162409-2906a7bf1935
	github.com/sirupsen/logrus => github.com/gravitational/logrus v0.10.1-0.20180402202453-dcdb95d728db
	k8s.io/api => k8s.io/api v0.19.14
	k8s.io/client-go => k8s.io/client-go v0.19.14
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
)
