package podage

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/internal"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var checkName string = "podage"

// QuerierSpec defines the Specification for a Querier.
type QuerierSpec struct {
}

// Querier defines the query and set of checks.
type Querier struct {
	client *kubernetes.Clientset
	Spec   *QuerierSpec `yaml:"spec"`
}

// AddtoRunner configures a runner with the Querier for this check.
func (querier *Querier) AddtoRunner(runner *internal.Runner) {
	runner.Queriers = append(runner.Queriers, querier)
	runner.Logger.WithFields(log.Fields{
		"check_name": checkName,
		"phase":      "add",
	}).Info("complete")
}

// NewQuerier returns a new configured Querier.
func NewQuerier(spec *QuerierSpec) (Querier, error) {
	out := Querier{
		Spec: spec,
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return out, err
	}
	out.client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return out, err
	}
	return out, nil
}

// Start runs the Querier.
func (q Querier) Start(cfg *internal.QuerierConfig) {
	cfg.Logger.WithFields(log.Fields{
		"check_name": checkName,
		"phase":      "add",
	}).Info(internal.CheckStartMsg)

	checkItem := internal.ReportItem{
		Name:   checkName,
		Status: "passed",
	}

	namespaces, err := q.client.CoreV1().Namespaces().List(cfg.Context, metav1.ListOptions{})
	if err != nil {
		checkItem.Status = "failed"
	}

	pods, err := q.client.CoreV1().Pods("").List(cfg.Context, metav1.ListOptions{})
	if err != nil {
		checkItem.Status = "failed"
	}
	threshold := time.Now().Sub(1 * time.Hour)
	for _, pod := range pods.Items {
		if (pod.ObjectMeta.CreationTimestamp.Before(threshold)) {
			// TODO: error
			checkItem.Items = append(checkItem.Items, internal.Item{
				Name: fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
				Status: "failed",
			})
			checkItem.Status = "failed"
		}
	}

	}

	cfg.Logger.WithFields(log.Fields{
		"component":  "check",
		"check_name": checkName,
		"phase":      "debug",
	}).Info(covered)

	cfg.Logger.WithFields(log.Fields{
		"component":  "check",
		"check_name": checkName,
		"phase":      "complete",
	}).Info(internal.CheckCompleteMsg)

	cfg.Results <- checkItem

	cfg.Logger.WithFields(log.Fields{
		"component":  "check",
		"check_name": checkName,
		"phase":      "write",
	}).Info(internal.CheckWriteMsg)
}
