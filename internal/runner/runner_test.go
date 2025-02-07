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

const binaryPath string = "bin/tenv-binaries"
const repositoryPath string = "test.out/runner-repository"

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

	// create directory for terraform/terragrunt binaries (to re-use them between tests)
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	err = os.MkdirAll(filepath.Join(cwd, binaryPath), 0755)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	err = os.RemoveAll(filepath.Join(cwd, binaryPath))
	Expect(err).NotTo(HaveOccurred())
})

func generateTestConfig() *config.Config {
	conf := config.TestConfig()
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	conf.Runner.RunnerBinaryPath = filepath.Join(cwd, binaryPath)
	conf.Runner.RepositoryPath = filepath.Join(cwd, repositoryPath)
	err = os.MkdirAll(conf.Runner.RepositoryPath, 0755)
	Expect(err).NotTo(HaveOccurred())
	return conf
}

// Remove the repository and the binaries directories
func cleanup(conf *config.Config) {
	_ = os.RemoveAll(conf.Runner.RepositoryPath)
}

func executeRunner(r *runner.Runner) error {
	r.Datastore = datastore.NewMockClient()
	r.Client = k8sClient
	var err error
	err = r.Init()
	if err != nil {
		return err
	}
	err = r.ExecInit()
	if err != nil {
		return err
	}
	return r.ExecAction()
}

var _ = Describe("Runner Tests", func() {
	var err error
	Describe("Nominal Case", Ordered, func() {
		Describe("End-to-End - When Runner is launched for running a Terraform plan", Ordered, func() {
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
		Describe("End-to-End - When Runner is launched for running a Terraform apply", Ordered, func() {
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
		Describe("End-to-End - When Runner is launched for running a Terragrunt plan", Ordered, func() {
			var conf *config.Config
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Runner.Action = "plan"
				conf.Runner.Layer.Name = "nominal-case-2"
				conf.Runner.Layer.Namespace = "default"
				conf.Runner.Run = "nominal-case-2-plan"

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
				err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "nominal-case-2"}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(layer.Annotations).To(HaveKey(annotations.LastPlanDate))
				Expect(layer.Annotations).To(HaveKey(annotations.LastPlanRun))
				Expect(layer.Annotations).To(HaveKey(annotations.LastPlanSum))
				Expect(layer.Annotations).To(HaveKey(annotations.LastPlanCommit))
			})
		})
		Describe("End-to-End - When Runner is launched for running Terragrunt apply", Ordered, func() {
			var conf *config.Config
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Runner.Action = "apply"
				conf.Runner.Layer.Name = "nominal-case-2"
				conf.Runner.Layer.Namespace = "default"
				conf.Runner.Run = "nominal-case-2-apply"

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
				err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "nominal-case-2"}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(layer.Annotations).To(HaveKey(annotations.LastApplyDate))
				Expect(layer.Annotations).To(HaveKey(annotations.LastApplySum))
				Expect(layer.Annotations).To(HaveKey(annotations.LastPlanCommit))
			})
		})
		Describe("When Hermitcrab is enabled", Ordered, func() {
			var conf *config.Config
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Hermitcrab.Enabled = true
				conf.Hermitcrab.URL = "http://hermitcrab.local"

				runner := runner.New(conf)
				err = runner.EnableHermitcrab()
			})
			AfterAll(func() {
				cleanup(conf)
				_ = os.Unsetenv("TF_CLI_CONFIG_FILE")
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should have created a network mirror configuration file that contains the URL", func() {
				content, err := os.ReadFile(filepath.Join(conf.Runner.RepositoryPath, "config.tfrc"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("http://hermitcrab.local"))
			})
			It("should have set the TF_CLI_CONFIG_FILE environment variable", func() {
				Expect(os.Getenv("TF_CLI_CONFIG_FILE")).NotTo(BeEmpty())
			})
		})
		Describe("When all resources are present (layer, run, repository)", Ordered, func() {
			var conf *config.Config
			var runnerInstance *runner.Runner
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Runner.Action = "plan"
				conf.Runner.Layer.Name = "nominal-case-1"
				conf.Runner.Layer.Namespace = "default"
				conf.Runner.Run = "nominal-case-1-plan"

				runnerInstance = runner.New(conf)
				runnerInstance.Client = k8sClient
				err = runnerInstance.GetResources()
			})
			AfterAll(func() {
				cleanup(conf)
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should have retrieved the layer", func() {
				Expect(runnerInstance.Layer).NotTo(BeNil())
			})
			It("should have retrieved the run", func() {
				Expect(runnerInstance.Run).NotTo(BeNil())
			})
			It("should have retrieved the repository", func() {
				Expect(runnerInstance.Repository).NotTo(BeNil())
			})
		})
		Describe("When binaries versions are specified in layer", Ordered, func() {
			var conf *config.Config
			var runnerInstance *runner.Runner
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Runner.Layer.Name = "nominal-case-2"
				conf.Runner.Layer.Namespace = "default"
				conf.Runner.Run = "nominal-case-2-plan"

				runnerInstance = runner.New(conf)
				runnerInstance.Client = k8sClient
				err = runnerInstance.Init()
			})
			AfterAll(func() {
				cleanup(conf)
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should have installed Terraform version 1.7.5", func() {
				_, err := os.Stat(filepath.Join(conf.Runner.RunnerBinaryPath, "Terraform", "1.7.5", "terraform"))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should have installed Terragrunt version 0.66.9", func() {
				_, err := os.Stat(filepath.Join(conf.Runner.RunnerBinaryPath, "Terragrunt", "0.66.9", "terragrunt"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Describe("When Terraform version is specified in the codebase", Ordered, func() {
			var conf *config.Config
			var runnerInstance *runner.Runner
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Runner.Layer.Name = "nominal-case-3"
				conf.Runner.Layer.Namespace = "default"
				conf.Runner.Run = "nominal-case-3-plan"

				runnerInstance = runner.New(conf)
				runnerInstance.Client = k8sClient
				err = runnerInstance.Init()
			})
			AfterAll(func() {
				cleanup(conf)
			})
			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("should have installed Terraform version 1.7.5", func() {
				_, err := os.Stat(filepath.Join(conf.Runner.RunnerBinaryPath, "Terraform", "1.7.5", "terraform"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
	Describe("Error Cases", Ordered, func() {
		Describe("When repository fails to fetch", Ordered, func() {
			var conf *config.Config
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Runner.Action = "plan"
				conf.Runner.Layer.Name = "error-case-1"
				conf.Runner.Layer.Namespace = "default"
				conf.Runner.Run = "error-case-1-plan"

				runner := runner.New(conf)
				err = executeRunner(runner)
			})
			AfterAll(func() {
				cleanup(conf)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
			It("should not have updated the TerraformLayer annotations", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "error-case-1"}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanDate))
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanSum))
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanCommit))
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanRun))
			})
		})
		Describe("When binaries version constraint are malformed", Ordered, func() {
			var conf *config.Config
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Runner.Action = "plan"
				conf.Runner.Layer.Name = "error-case-2"
				conf.Runner.Layer.Namespace = "default"
				conf.Runner.Run = "error-case-2-plan"

				runner := runner.New(conf)
				err = executeRunner(runner)
			})
			AfterAll(func() {
				cleanup(conf)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
			It("should not have updated the TerraformLayer annotations", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "error-case-2"}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanDate))
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanSum))
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanCommit))
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanRun))
			})
		})
		Describe("When binaries version does not exist", Ordered, func() {
			var conf *config.Config
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Runner.Action = "plan"
				conf.Runner.Layer.Name = "error-case-3"
				conf.Runner.Layer.Namespace = "default"
				conf.Runner.Run = "error-case-3-plan"

				runner := runner.New(conf)
				err = executeRunner(runner)
			})
			AfterAll(func() {
				cleanup(conf)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
			It("should not have updated the TerraformLayer annotations", func() {
				layer := &configv1alpha1.TerraformLayer{}
				err := k8sClient.Get(context.TODO(), client.ObjectKey{Namespace: "default", Name: "error-case-3"}, layer)
				Expect(err).NotTo(HaveOccurred())
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanDate))
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanSum))
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanCommit))
				Expect(layer.Annotations).NotTo(HaveKey(annotations.LastPlanRun))
			})
		})
		Describe("When linked resources do not exist", Ordered, func() {
			var conf *config.Config
			var runnerInstance *runner.Runner
			BeforeAll(func() {
				conf = generateTestConfig()
				conf.Runner.Action = "plan"
				conf.Runner.Layer.Name = "non-existent-layer"
				conf.Runner.Layer.Namespace = "default"
				conf.Runner.Run = "non-existent-run"

				runnerInstance = runner.New(conf)
				err = executeRunner(runnerInstance)
			})
			AfterAll(func() {
				cleanup(conf)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
			It("should have a nil layer", func() {
				Expect(runnerInstance.Layer).To(BeNil())
			})
			It("should have a nil run", func() {
				Expect(runnerInstance.Run).To(BeNil())
			})
			It("should have a nil repository", func() {
				Expect(runnerInstance.Repository).To(BeNil())
			})
		})
	})
})
