/*
Copyright 2022.

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

package controllers

import (
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func TestConditions(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Conditions Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
})

var _ = AfterSuite(func() {
})

var _ = Describe("TerraformLayer", func() {
	var t *configv1alpha1.TerraformLayer
	var cache Cache

	BeforeEach(func() {
		t = &configv1alpha1.TerraformLayer{
			Spec: configv1alpha1.TerraformLayerSpec{
				Path:   "/test",
				Branch: "main",
				Repository: configv1alpha1.TerraformLayerRepository{
					Name:      "test-repository",
					Namespace: "default",
				},
			},
		}
		cache = newMemoryCache()
	})

	Describe("TerraformRunningCondition", func() {
		var condition TerraformRunningCondition
		BeforeEach(func() {
			condition = TerraformRunningCondition{}
		})
		Context("with lock in cache", func() {
			It("should return true", func() {
				cache.Set(PrefixLock+computeHash(t.Spec.Repository.Name, t.Spec.Repository.Namespace, t.Spec.Path), []byte{1}, 0)
				Expect(condition.Evaluate(cache, t)).To(Equal(true))
			})
		})
		Context("with lock not in cache", func() {
			It("should return false", func() {
				Expect(condition.Evaluate(cache, t)).To(Equal(false))
			})
		})
	})
	Describe("TerraformPlanArtifactCondition", func() {
		var condition TerraformPlanArtifactCondition
		BeforeEach(func() {
			condition = TerraformPlanArtifactCondition{}
		})
		Context("with no last timestamp in cache", func() {
			It("should return false", func() {
				Expect(condition.Evaluate(cache, t)).To(Equal(false))
			})
		})
		Context("with last timestamp in cache < 20min", func() {
			It("should return true", func() {
				cache.Set(PrefixLastPlanDate+computeHash(t.Spec.Repository.Name, t.Spec.Repository.Namespace, t.Spec.Path, t.Spec.Branch),
					[]byte(strconv.Itoa(int((time.Now().Add(-5 * time.Minute)).Unix()))), 0)
				Expect(condition.Evaluate(cache, t)).To(Equal(true))
			})
		})
		Context("with last timestamp in cache > 20min", func() {
			It("should return false", func() {
				cache.Set(PrefixLastPlanDate+computeHash(t.Spec.Repository.Name, t.Spec.Repository.Namespace, t.Spec.Path, t.Spec.Branch),
					[]byte(strconv.Itoa(int(time.Now().Add(-25*time.Minute).Unix()))), 0)
				Expect(condition.Evaluate(cache, t)).To(Equal(false))
			})
		})
	})
	// Describe("TerraformPlanArtifactCondition", func() {
	// 	var runningCondition TerraformRunningCondition
	// 	BeforeEach(func() {
	// 		runningCondition = TerraformRunningCondition{}
	// 	})
	// 	Context("with lock in cache", func() {
	// 		It("should return true", func() {
	// 			cache.Set(PrefixLock+computeHash(t.Spec.Repository.Name, t.Spec.Repository.Namespace, t.Spec.Path), []byte{1}, 0)
	// 			Expect(runningCondition.Evaluate(cache, t)).To(Equal(true))
	// 		})
	// 	})
	// 	Context("with lock not in cache", func() {
	// 		It("should return false", func() {
	// 			Expect(runningCondition.Evaluate(cache, t)).To(Equal(false))
	// 		})
	// 	})
	// })
	// Describe("TerraformPlanArtifactCondition", func() {
	// 	var runningCondition TerraformRunningCondition
	// 	BeforeEach(func() {
	// 		runningCondition = TerraformRunningCondition{}
	// 	})
	// 	Context("with lock in cache", func() {
	// 		It("should return true", func() {
	// 			cache.Set(PrefixLock+computeHash(t.Spec.Repository.Name, t.Spec.Repository.Namespace, t.Spec.Path), []byte{1}, 0)
	// 			Expect(runningCondition.Evaluate(cache, t)).To(Equal(true))
	// 		})
	// 	})
	// 	Context("with lock not in cache", func() {
	// 		It("should return false", func() {
	// 			Expect(runningCondition.Evaluate(cache, t)).To(Equal(false))
	// 		})
	// 	})
	// })
})
