/*
Copyright 2022 Hiroki Okui.

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

package internal

import (
	"fmt"
	"github.com/hrk091/retryable-pipeline/api/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"path"
	"runtime"
)

func NewRetryablePipelineRunTestData(name string) *v1alpha1.RetryablePipelineRun {
	buf, err := ioutil.ReadFile(fmt.Sprintf("%s/testdata/%s.yaml", pkgDir(), name))
	if err != nil {
		panic(err)
	}
	var o v1alpha1.RetryablePipelineRun
	if err := yaml.Unmarshal(buf, &o); err != nil {
		panic(err)
	}
	return &o
}

func NewPipelineTestData(name string) *v1beta1.Pipeline {
	buf, err := ioutil.ReadFile(fmt.Sprintf("%s/testdata/%s.yaml", pkgDir(), name))
	if err != nil {
		panic(err)
	}
	var o v1beta1.Pipeline
	if err := yaml.Unmarshal(buf, &o); err != nil {
		panic(err)
	}
	return &o
}

func NewPipelineRunTestData(name string) *v1beta1.PipelineRun {
	buf, err := ioutil.ReadFile(fmt.Sprintf("%s/testdata/%s.yaml", pkgDir(), name))
	if err != nil {
		panic(err)
	}
	var o v1beta1.PipelineRun
	if err := yaml.Unmarshal(buf, &o); err != nil {
		panic(err)
	}
	return &o
}

func NewTaskTestData(name string) *v1beta1.Task {
	buf, err := ioutil.ReadFile(fmt.Sprintf("%s/testdata/%s.yaml", pkgDir(), name))
	if err != nil {
		panic(err)
	}
	var o v1beta1.Task
	if err := yaml.Unmarshal(buf, &o); err != nil {
		panic(err)
	}
	return &o
}

func NewPvcTestData(name string) *corev1.PersistentVolumeClaim {
	buf, err := ioutil.ReadFile(fmt.Sprintf("%s/testdata/%s.yaml", pkgDir(), name))
	if err != nil {
		panic(err)
	}
	var o corev1.PersistentVolumeClaim
	if err := yaml.Unmarshal(buf, &o); err != nil {
		panic(err)
	}
	return &o
}

func pkgDir() string {
	_, b, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(b))
}
