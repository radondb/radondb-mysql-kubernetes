package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultHost           = "sample-leader"
	DefaultUser           = "radondb_usr"
	DefaultPassword       = "RadonDB@123"
	DefaultDB             = "radondb"
	DefaultTables         = 4
	DefaultTableSize      = 2000
	DefaultReportInterval = 10
	DefaultTime           = 9999
	DefaultThreads        = 8
)

type SysbenchOptions struct {
	Host           string
	User           string
	Password       string
	DB             string
	Tables         int
	TableSize      int
	Threads        int
	ReportInterval int
	Time           int
}

func NewDefaultSysbenchOptions() *SysbenchOptions {
	return &SysbenchOptions{
		Host:           DefaultHost,
		User:           DefaultUser,
		Password:       DefaultPassword,
		DB:             DefaultDB,
		Tables:         DefaultTables,
		TableSize:      DefaultTableSize,
		Threads:        DefaultThreads,
		ReportInterval: DefaultReportInterval,
		Time:           DefaultTime,
	}
}

func sysbenchContainer(o *SysbenchOptions, phase string) *corev1.Container {
	script := "/usr/share/sysbench/oltp_common.lua"
	if phase == "run" {
		script = "/usr/share/sysbench/oltp_read_write.lua"
	}
	return &corev1.Container{
		Name:    "sysbench",
		Image:   "severalnines/sysbench",
		Command: []string{"sysbench"},
		Args: []string{
			"--db-driver=mysql",
			"--mysql-port=3306",
			fmt.Sprintf("--mysql-host=%s", o.Host),
			fmt.Sprintf("--mysql-user=%s", o.User),
			fmt.Sprintf("--mysql-password=%s", o.Password),
			fmt.Sprintf("--mysql-db=%s", o.DB),
			fmt.Sprintf("--tables=%d", o.Tables),
			fmt.Sprintf("--table-size=%d", o.TableSize),
			fmt.Sprintf("--threads=%d", o.Threads),
			fmt.Sprintf("--time=%d", o.Time),
			fmt.Sprintf("--report-interval=%d", o.ReportInterval),
			script,
			phase,
		},
	}
}

func (f *Framework) PrepareData(o *SysbenchOptions) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sysbench-prepare",
			Namespace: f.Namespace.Name,
		},
	}
	job.Spec.Template.Spec.Containers = []corev1.Container{*sysbenchContainer(o, "prepare")}
	job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
	Expect(f.Client.Create(context.TODO(), job, &client.CreateOptions{})).Should(Succeed())
	k8s.WaitUntilJobSucceed(f.t, f.kubectlOptions, job.Name, 12, 5*time.Second)
	Expect(f.Client.Delete(context.TODO(), job, &client.DeleteOptions{})).Should(Succeed())
}

func (f *Framework) RunSysbench(o *SysbenchOptions) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sysbench-run",
			Namespace: f.Namespace.Name,
		},
	}
	pod.Spec.Containers = []corev1.Container{*sysbenchContainer(o, "run")}
	Expect(f.Client.Create(context.TODO(), pod, &client.CreateOptions{})).Should(Succeed())
	k8s.WaitUntilPodAvailable(f.t, f.kubectlOptions, pod.Name, 12, 5*time.Second)
}

func (f *Framework) CleanUpSysbench() {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sysbench-run",
			Namespace: f.Namespace.Name,
		},
	}
	Expect(f.Client.Delete(context.TODO(), pod, &client.DeleteAllOfOptions{})).Should(Succeed())
}
