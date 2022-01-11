module github.com/radondb/radondb-mysql-kubernetes

go 1.16

require (
	bou.ke/monkey v1.0.2
	github.com/blang/semver v3.5.1+incompatible
	github.com/go-ini/ini v1.62.0
	github.com/go-logr/logr v0.4.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-test/deep v1.0.7
	github.com/golang/glog v1.0.0
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365
	github.com/imdario/mergo v0.3.12
	github.com/onsi/ginkgo/v2 v2.0.0
	github.com/onsi/gomega v1.17.0
	github.com/presslabs/controller-util v0.3.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	golang.org/x/tools v0.1.8-0.20211028023602-8de2a7fd1736 // indirect
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	k8s.io/klog/v2 v2.8.0
	sigs.k8s.io/controller-runtime v0.9.5
)
