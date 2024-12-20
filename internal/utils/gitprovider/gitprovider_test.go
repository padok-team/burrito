package gitprovider_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/padok-team/burrito/internal/utils/gitprovider"
	"github.com/padok-team/burrito/internal/utils/gitprovider/types"
)

func TestGitProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitProvider Suite")
}

var _ = Describe("GitProvider", func() {
	Context("ListAvailable", func() {
		Describe("When checking available providers with multiple GitHub credentials", func() {
			var (
				config       types.Config
				providers    []string
				err          error
				capabilities []string
			)

			BeforeEach(func() {
				config = types.Config{
					// GitHub App credentials
					AppID:             123,
					AppInstallationID: 456,
					AppPrivateKey:     "test-key",
					// Basic auth credentials
					Username: "user",
					Password: "pass",
					URL:      "https://github.com/org/repo",
				}
				capabilities = []string{types.Capabilities.Clone, types.Capabilities.Comment}
			})

			JustBeforeEach(func() {
				providers, err = gitprovider.ListAvailable(config, capabilities)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should list github as first provider", func() {
				Expect(providers[0]).To(Equal("github"))
			})
		})

		Describe("When checking available providers with multiple GitLab credentials", func() {
			var (
				config       types.Config
				providers    []string
				err          error
				capabilities []string
			)

			BeforeEach(func() {
				config = types.Config{
					GitLabToken: "test-token",
					Username:    "user",
					Password:    "pass",
					URL:         "https://gitlab.com/org/repo",
				}
				capabilities = []string{types.Capabilities.Clone, types.Capabilities.Comment}
			})

			JustBeforeEach(func() {
				providers, err = gitprovider.ListAvailable(config, capabilities)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should list gitlab as first provider", func() {
				Expect(providers[0]).To(Equal("gitlab"))
			})
		})

		Describe("When checking available providers with GitHub token", func() {
			var (
				config       types.Config
				providers    []string
				err          error
				capabilities []string
			)

			BeforeEach(func() {
				config = types.Config{
					GitHubToken: "test-token",
					URL:         "https://github.com/org/repo",
				}
				capabilities = []string{types.Capabilities.Clone, types.Capabilities.Comment}
			})

			JustBeforeEach(func() {
				providers, err = gitprovider.ListAvailable(config, capabilities)
			})

			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should list github as available provider", func() {
				Expect(providers).To(ContainElement("github"))
			})

			It("should not list gitlab as available provider", func() {
				Expect(providers).NotTo(ContainElement("gitlab"))
			})

			It("should not list standard as available provider", func() {
				Expect(providers).NotTo(ContainElement("standard"))
			})
		})

		Describe("When checking available providers with GitLab token", func() {
			var (
				config       types.Config
				providers    []string
				err          error
				capabilities []string
			)

			BeforeEach(func() {
				config = types.Config{
					GitLabToken: "test-token",
					URL:         "https://gitlab.com/org/repo",
				}
				capabilities = []string{types.Capabilities.Clone, types.Capabilities.Comment}
			})

			JustBeforeEach(func() {
				providers, err = gitprovider.ListAvailable(config, capabilities)
			})

			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should list gitlab as available provider", func() {
				Expect(providers).To(ContainElement("gitlab"))
			})

			It("should not list github as available provider", func() {
				Expect(providers).NotTo(ContainElement("github"))
			})
		})

		Describe("When checking available providers with only clone capability", func() {
			var (
				config       types.Config
				providers    []string
				err          error
				capabilities []string
			)

			BeforeEach(func() {
				config = types.Config{
					Username: "user",
					Password: "pass",
					URL:      "https://github.com/org/repo",
				}
				capabilities = []string{types.Capabilities.Clone}
			})

			JustBeforeEach(func() {
				providers, err = gitprovider.ListAvailable(config, capabilities)
			})

			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should list standard as available provider", func() {
				Expect(providers).To(ContainElement("standard"))
			})
		})
	})

	Context("New", func() {
		Describe("When creating a new provider with GitHub token", func() {
			var (
				config       types.Config
				provider     types.Provider
				err          error
				capabilities []string
			)

			BeforeEach(func() {
				config = types.Config{
					GitHubToken: "test-token",
					URL:         "https://github.com/org/repo",
				}
				capabilities = []string{types.Capabilities.Clone, types.Capabilities.Comment}
			})

			JustBeforeEach(func() {
				provider, err = gitprovider.New(config, capabilities)
			})

			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a non-nil provider", func() {
				Expect(provider).NotTo(BeNil())
			})
		})

		Describe("When creating a new provider with invalid configuration", func() {
			var (
				config       types.Config
				provider     types.Provider
				err          error
				capabilities []string
			)

			BeforeEach(func() {
				config = types.Config{
					URL: "https://github.com/org/repo",
				}
				capabilities = []string{types.Capabilities.Clone, types.Capabilities.Comment}
			})

			JustBeforeEach(func() {
				provider, err = gitprovider.New(config, capabilities)
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})

			It("should return a nil provider", func() {
				Expect(provider).To(BeNil())
			})
		})
	})

	Context("NewWithName", func() {
		Describe("When creating a provider with a specific name", func() {
			var (
				config   types.Config
				provider types.Provider
				err      error
			)

			BeforeEach(func() {
				config = types.Config{
					EnableMock: true,
				}
			})

			JustBeforeEach(func() {
				provider, err = gitprovider.NewWithName(config, "mock")
			})

			It("should not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a non-nil provider", func() {
				Expect(provider).NotTo(BeNil())
			})
		})

		Describe("When creating a provider with an invalid name", func() {
			var (
				config   types.Config
				provider types.Provider
				err      error
			)

			BeforeEach(func() {
				config = types.Config{}
			})

			JustBeforeEach(func() {
				provider, err = gitprovider.NewWithName(config, "invalid")
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})

			It("should return a nil provider", func() {
				Expect(provider).To(BeNil())
			})
		})
	})
})
