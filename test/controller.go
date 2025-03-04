/*
Copyright 2019 The Tekton Authors
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

package test

import (
	"context"
	"testing"

	// Link in the fakes so they get injected into injection.Fake
	fakekubeclient "github.com/knative/pkg/injection/clients/kubeclient/fake"
	fakepodinformer "github.com/knative/pkg/injection/informers/kubeinformers/corev1/pod/fake"
	fakepipelineclient "github.com/tektoncd/pipeline/pkg/client/injection/client/fake"
	fakeclustertaskinformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1alpha1/clustertask/fake"
	fakepipelineinformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1alpha1/pipeline/fake"
	fakeresourceinformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1alpha1/pipelineresource/fake"
	fakepipelineruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1alpha1/pipelinerun/fake"
	faketaskinformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1alpha1/task/fake"
	faketaskruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1alpha1/taskrun/fake"

	"github.com/knative/pkg/controller"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	fakepipelineclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/fake"
	informersv1alpha1 "github.com/tektoncd/pipeline/pkg/client/informers/externalversions/pipeline/v1alpha1"
	"go.uber.org/zap/zaptest/observer"
	corev1 "k8s.io/api/core/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
)

// GetLogMessages returns a list of all string logs in logs.
func GetLogMessages(logs *observer.ObservedLogs) []string {
	messages := []string{}
	for _, l := range logs.All() {
		messages = append(messages, l.Message)
	}
	return messages
}

// Data represents the desired state of the system (i.e. existing resources) to seed controllers
// with.
type Data struct {
	PipelineRuns      []*v1alpha1.PipelineRun
	Pipelines         []*v1alpha1.Pipeline
	TaskRuns          []*v1alpha1.TaskRun
	Tasks             []*v1alpha1.Task
	ClusterTasks      []*v1alpha1.ClusterTask
	PipelineResources []*v1alpha1.PipelineResource
	Pods              []*corev1.Pod
	Namespaces        []*corev1.Namespace
}

// Clients holds references to clients which are useful for reconciler tests.
type Clients struct {
	Pipeline *fakepipelineclientset.Clientset
	Kube     *fakekubeclientset.Clientset
}

// Informers holds references to informers which are useful for reconciler tests.
type Informers struct {
	PipelineRun      informersv1alpha1.PipelineRunInformer
	Pipeline         informersv1alpha1.PipelineInformer
	TaskRun          informersv1alpha1.TaskRunInformer
	Task             informersv1alpha1.TaskInformer
	ClusterTask      informersv1alpha1.ClusterTaskInformer
	PipelineResource informersv1alpha1.PipelineResourceInformer
	Pod              coreinformers.PodInformer
}

// TestAssets holds references to the controller, logs, clients, and informers.
type TestAssets struct {
	Controller *controller.Impl
	Logs       *observer.ObservedLogs
	Clients    Clients
}

// SeedTestData returns Clients and Informers populated with the
// given Data.
func SeedTestData(t *testing.T, ctx context.Context, d Data) (Clients, Informers) {
	c := Clients{
		Kube:     fakekubeclient.Get(ctx),
		Pipeline: fakepipelineclient.Get(ctx),
	}

	i := Informers{
		PipelineRun:      fakepipelineruninformer.Get(ctx),
		Pipeline:         fakepipelineinformer.Get(ctx),
		TaskRun:          faketaskruninformer.Get(ctx),
		Task:             faketaskinformer.Get(ctx),
		ClusterTask:      fakeclustertaskinformer.Get(ctx),
		PipelineResource: fakeresourceinformer.Get(ctx),
		Pod:              fakepodinformer.Get(ctx),
	}

	for _, pr := range d.PipelineRuns {
		if err := i.PipelineRun.Informer().GetIndexer().Add(pr); err != nil {
			t.Fatal(err)
		}
		if _, err := c.Pipeline.TektonV1alpha1().PipelineRuns(pr.Namespace).Create(pr); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range d.Pipelines {
		if err := i.Pipeline.Informer().GetIndexer().Add(p); err != nil {
			t.Fatal(err)
		}
		if _, err := c.Pipeline.TektonV1alpha1().Pipelines(p.Namespace).Create(p); err != nil {
			t.Fatal(err)
		}
	}
	for _, tr := range d.TaskRuns {
		if err := i.TaskRun.Informer().GetIndexer().Add(tr); err != nil {
			t.Fatal(err)
		}
		if _, err := c.Pipeline.TektonV1alpha1().TaskRuns(tr.Namespace).Create(tr); err != nil {
			t.Fatal(err)
		}
	}
	for _, ta := range d.Tasks {
		if err := i.Task.Informer().GetIndexer().Add(ta); err != nil {
			t.Fatal(err)
		}
		if _, err := c.Pipeline.TektonV1alpha1().Tasks(ta.Namespace).Create(ta); err != nil {
			t.Fatal(err)
		}
	}
	for _, ct := range d.ClusterTasks {
		if err := i.ClusterTask.Informer().GetIndexer().Add(ct); err != nil {
			t.Fatal(err)
		}
		if _, err := c.Pipeline.TektonV1alpha1().ClusterTasks().Create(ct); err != nil {
			t.Fatal(err)
		}
	}
	for _, r := range d.PipelineResources {
		if err := i.PipelineResource.Informer().GetIndexer().Add(r); err != nil {
			t.Fatal(err)
		}
		if _, err := c.Pipeline.TektonV1alpha1().PipelineResources(r.Namespace).Create(r); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range d.Pods {
		if err := i.Pod.Informer().GetIndexer().Add(p); err != nil {
			t.Fatal(err)
		}
		if _, err := c.Kube.CoreV1().Pods(p.Namespace).Create(p); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range d.Namespaces {
		if _, err := c.Kube.CoreV1().Namespaces().Create(n); err != nil {
			t.Fatal(err)
		}
	}
	c.Pipeline.ClearActions()
	c.Kube.ClearActions()
	return c, i
}
