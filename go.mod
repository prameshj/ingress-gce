module k8s.io/ingress-gce

go 1.12

require (
	github.com/GoogleCloudPlatform/k8s-cloud-provider v0.0.0-20190612171043-2e19bb35a278
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/beorn7/perks v1.0.0 // indirect
	github.com/emicklei/go-restful v2.9.3+incompatible // indirect
	github.com/evanphx/json-patch v4.1.0+incompatible // indirect
	github.com/go-openapi/spec v0.19.0
	github.com/go-openapi/swag v0.19.0 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/kr/pretty v0.1.0
	github.com/mailru/easyjson v0.0.0-20190403194419-1ea4449da983 // indirect
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/common v0.3.0 // indirect
	github.com/prometheus/procfs v0.0.0-20190416084830-8368d24ba045 // indirect
	github.com/spf13/pflag v1.0.3
	golang.org/x/crypto v0.0.0-20190422183909-d864b10871cd // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/api v0.6.1-0.20190607001116-5213b8090861
	gopkg.in/gcfg.v1 v1.2.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/apiserver v0.0.0
	k8s.io/client-go v0.0.0
	k8s.io/cloud-provider v0.0.0
	k8s.io/component-base v0.0.0
	k8s.io/klog v0.3.1
	k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/kubernetes v1.15.0
	k8s.io/legacy-cloud-providers v0.0.0
	k8s.io/utils v0.0.0-20190221042446-c2654d5206da
)

replace (
	cloud.google.com/go => cloud.google.com/go v0.37.4
	github.com/GoogleCloudPlatform/k8s-cloud-provider => github.com/GoogleCloudPlatform/k8s-cloud-provider v0.0.0-20190612171043-2e19bb35a278
	github.com/PuerkitoBio/purell => github.com/PuerkitoBio/purell v1.1.1
	github.com/beorn7/perks => github.com/beorn7/perks v1.0.0
	github.com/emicklei/go-restful => github.com/emicklei/go-restful v2.9.3+incompatible
	github.com/evanphx/json-patch => github.com/evanphx/json-patch v4.1.0+incompatible
	github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.0
	github.com/go-openapi/swag => github.com/go-openapi/swag v0.19.0
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.2.1
	github.com/golang/groupcache => github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.1
	github.com/google/gofuzz => github.com/google/gofuzz v1.0.0
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.2.0
	github.com/hashicorp/golang-lru => github.com/hashicorp/golang-lru v0.5.1
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.7
	github.com/json-iterator/go => github.com/json-iterator/go v1.1.6
	github.com/kr/pretty => github.com/kr/pretty v0.1.0
	github.com/mailru/easyjson => github.com/mailru/easyjson v0.0.0-20190403194419-1ea4449da983
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/client_model => github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/prometheus/common => github.com/prometheus/common v0.3.0
	github.com/prometheus/procfs => github.com/prometheus/procfs v0.0.0-20190416084830-8368d24ba045
	github.com/spf13/pflag => github.com/spf13/pflag v1.0.3
	go.opencensus.io => go.opencensus.io v0.20.2
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20190422183909-d864b10871cd
	golang.org/x/net => golang.org/x/net v0.0.0-20190420063019-afa5a82059c6
	golang.org/x/oauth2 => golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190422165155-953cdadca894
	golang.org/x/time => golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/api => google.golang.org/api v0.6.1-0.20190607001116-5213b8090861
	gopkg.in/gcfg.v1 => gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/inf.v0 => gopkg.in/inf.v0 v0.9.1
	gopkg.in/warnings.v0 => gopkg.in/warnings.v0 v0.1.2
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.2
	k8s.io/api => k8s.io/api v0.0.0-20190620085002-8f739060a0b3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085550-3a2f62f126c9
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190620085203-5d32fb3b42f4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190620085659-429467d76d0e
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085041-d697df55dbe9
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190620090041-1a7e1f6630cd
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190620090010-a60497bb9ffa
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190620085131-4cd66be69262
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190620090114-816aa063c73d
	k8s.io/klog => k8s.io/klog v0.3.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190620085316-c835efc41000
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190620085943-52018c8ce3c1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190620085811-cc0b23ba60a9
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190620085909-5dfb14b3a101
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190620085837-98477dc0c87c
	k8s.io/kubernetes => k8s.io/kubernetes v1.15.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190620090159-a9e4f3cb5bf3
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190620085627-5b02f62e9559
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190620085357-8191e314a1f7
)
