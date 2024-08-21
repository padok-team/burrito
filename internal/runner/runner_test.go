package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/runner"
	utils "github.com/padok-team/burrito/internal/testing"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestRunner(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Runner Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("../..", "manifests", "crds")},
		ErrorIfCRDPathMissing: true,
	}
	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = configv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	Expect(err).NotTo(HaveOccurred())
	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	utils.LoadResources(k8sClient, "testdata")

	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func generateTestConfig() *config.Config {
	conf := config.TestConfig()
	cwd, _ := os.Getwd()
	conf.Runner.RunnerBinaryPath = filepath.Join(cwd, "bin", "tenv-binaries")
	conf.Runner.RepositoryPath = filepath.Join(cwd, "test.out", "runner-repository")
	_ = os.MkdirAll(conf.Runner.RepositoryPath, 0755)
	_ = os.MkdirAll(conf.Runner.RunnerBinaryPath, 0755)

	return conf
}

// Remove the repository and the binaries directories
func cleanup(conf *config.Config) {
	_ = os.RemoveAll(conf.Runner.RepositoryPath)
	_ = os.RemoveAll(conf.Runner.RunnerBinaryPath)
}

func executeRunner(r *runner.Runner) error {
	r.Datastore = datastore.NewMockClient()
	r.Client = k8sClient
	err := r.Init()
	if err != nil {
		return err
	}
	err = r.ExecInit()
	if err != nil {
		return err
	}
	return r.ExecAction()
}

var _ = Describe("End-to-End Runner Tests", func() {
	var err error
	Describe("Nominal Case", func() {
		Describe("When Runner is launched for running a Plan", Ordered, func() {
			var conf *config.Config
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Runner.Action = "plan"
				conf.Runner.Layer.Name = "nominal-case-1"
				conf.Runner.Layer.Namespace = "default"
				conf.Runner.Run = "nominal-case-1-plan"

				runner := runner.New(conf)
				err = executeRunner(runner)
			})
			AfterAll(func() {
				cleanup(conf)
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should have updated the TerraformLayer annotations", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "nominal-case-1"}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(layer.Annotations).To(HaveKey(annotations.LastPlanDate))
				Expect(layer.Annotations).To(HaveKey(annotations.LastPlanRun))
				Expect(layer.Annotations).To(HaveKey(annotations.LastPlanSum))
				Expect(layer.Annotations).To(HaveKey(annotations.LastPlanCommit))
			})
		})
		Describe("When Runner is launched for running an Apply", Ordered, func() {
			var conf *config.Config
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Runner.Action = "apply"
				conf.Runner.Layer.Name = "nominal-case-1"
				conf.Runner.Layer.Namespace = "default"
				conf.Runner.Run = "nominal-case-1-apply"

				runner := runner.New(conf)
				err = executeRunner(runner)
			})
			AfterAll(func() {
				cleanup(conf)
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should have updated the TerraformLayer annotations", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "nominal-case-1"}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(layer.Annotations).To(HaveKey(annotations.LastApplyDate))
				Expect(layer.Annotations).To(HaveKey(annotations.LastApplySum))
				Expect(layer.Annotations).To(HaveKey(annotations.LastPlanCommit))
			})
		})

	})
})
