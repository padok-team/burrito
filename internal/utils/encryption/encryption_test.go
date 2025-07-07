package encryption

import (
	"bytes"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"
)

const (
	benchmarkSize       = 51200 // 50KB
	benchmarkIterations = 100
)

func TestEncryption(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Encryption Suite")
}

var _ = Describe("Encryptor", func() {
	DescribeTable("Encrypt/Decrypt operations",
		func(key string, plaintext []byte) {
			encryptor, err := NewEncryptor(key)
			Expect(err).NotTo(HaveOccurred(), "failed to create encryptor")

			By("encrypting the plaintext")
			ciphertext, err := encryptor.Encrypt(plaintext)
			Expect(err).NotTo(HaveOccurred(), "encryption failed")

			if len(plaintext) > 0 {
				By("ensuring ciphertext is different from plaintext")
				Expect(bytes.Equal(plaintext, ciphertext)).To(BeFalse(), "ciphertext should be different from plaintext")
			}

			By("decrypting the ciphertext")
			decrypted, err := encryptor.Decrypt(ciphertext)
			Expect(err).NotTo(HaveOccurred(), "decryption failed")

			By("verifying the decrypted data matches the original")
			Expect(bytes.Equal(plaintext, decrypted)).To(BeTrue(), "decrypted data should match original")
		},
		Entry("small text", "test-key-123", []byte("hello world")),
		Entry("terraform plan json", "terraform-plan-key", []byte(`{"format_version":"1.1","terraform_version":"1.0.0","planned_values":{}}`)),
		Entry("large binary data", "binary-key", make([]byte, benchmarkSize)),
		Entry("empty data", "empty-key", []byte{}),
		Entry("unicode text", "unicode-key", []byte("Hello 世界 🌍")),
	)

	Describe("Different Keys", func() {
		It("should fail to decrypt with a different key", func() {
			plaintext := []byte("sensitive terraform plan data")

			encryptor1, err := NewEncryptor("key1")
			Expect(err).NotTo(HaveOccurred(), "failed to create encryptor1")

			encryptor2, err := NewEncryptor("key2")
			Expect(err).NotTo(HaveOccurred(), "failed to create encryptor2")

			By("encrypting with first key")
			ciphertext, err := encryptor1.Encrypt(plaintext)
			Expect(err).NotTo(HaveOccurred(), "encryption failed")

			By("trying to decrypt with second key (should fail)")
			_, err = encryptor2.Decrypt(ciphertext)
			Expect(err).To(HaveOccurred(), "decryption should fail with different key")
		})
	})

	DescribeTable("Key Hashing",
		func(key string) {
			plaintext := []byte("test data for key hashing")

			encryptor, err := NewEncryptor(key)
			Expect(err).NotTo(HaveOccurred(), "failed to create encryptor")

			ciphertext, err := encryptor.Encrypt(plaintext)
			Expect(err).NotTo(HaveOccurred(), "encryption failed")

			decrypted, err := encryptor.Decrypt(ciphertext)
			Expect(err).NotTo(HaveOccurred(), "decryption failed")

			Expect(bytes.Equal(plaintext, decrypted)).To(BeTrue(), "decrypted data should match original")
		},
		Entry("short key", "short"),
		Entry("medium key", "this-is-a-medium-length-key"),
		Entry("long key", "this-is-a-very-long-key-that-exceeds-32-bytes-and-should-be-hashed-properly"),
	)

	Describe("Nonce Uniqueness", func() {
		It("should generate different ciphertexts for same plaintext", func() {
			encryptor, err := NewEncryptor("test-key")
			Expect(err).NotTo(HaveOccurred(), "failed to create encryptor")

			plaintext := []byte("same plaintext")

			By("encrypting the same plaintext multiple times")
			ciphertexts := make([][]byte, 10)
			for i := 0; i < 10; i++ {
				ciphertext, err := encryptor.Encrypt(plaintext)
				Expect(err).NotTo(HaveOccurred(), "encryption %d failed", i)
				ciphertexts[i] = ciphertext
			}

			By("verifying all ciphertexts are different (due to random nonces)")
			for i := 0; i < len(ciphertexts); i++ {
				for j := i + 1; j < len(ciphertexts); j++ {
					Expect(bytes.Equal(ciphertexts[i], ciphertexts[j])).To(BeFalse(), "ciphertexts %d and %d should be different", i, j)
				}
			}
		})
	})

	DescribeTable("Invalid Ciphertext",
		func(ciphertext []byte) {
			encryptor, err := NewEncryptor("test-key")
			Expect(err).NotTo(HaveOccurred(), "failed to create encryptor")

			_, err = encryptor.Decrypt(ciphertext)
			Expect(err).To(HaveOccurred(), "decryption should fail for invalid ciphertext")
		},
		Entry("too short", []byte("short")),
		Entry("corrupted data", make([]byte, 100)), // 100 bytes of zeros
	)

	Describe("NewEncryptor", func() {
		It("should return an error for empty key", func() {
			_, err := NewEncryptor("")
			Expect(err).To(HaveOccurred(), "NewEncryptor should return an error for empty key")

			expectedError := "encryption key cannot be empty"
			Expect(err.Error()).To(Equal(expectedError), "expected error message should match")
		})
	})

	Describe("Performance Benchmarks", func() {
		var encryptor *Encryptor
		var plaintext []byte
		var ciphertext []byte

		BeforeEach(func() {
			var err error
			encryptor, err = NewEncryptor("benchmark-key")
			Expect(err).NotTo(HaveOccurred(), "failed to create encryptor")

			plaintext = make([]byte, benchmarkSize)
			ciphertext, err = encryptor.Encrypt(plaintext)
			Expect(err).NotTo(HaveOccurred(), "failed to encrypt for benchmark setup")
		})

		It("should encrypt data efficiently", func() {
			experiment := gmeasure.NewExperiment("Encryption Performance")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("encryption", func() {
					_, err := encryptor.Encrypt(plaintext)
					Expect(err).NotTo(HaveOccurred())
				})
			}, gmeasure.SamplingConfig{N: benchmarkIterations})

			Expect(experiment.GetStats("encryption").DurationFor(gmeasure.StatMean)).To(BeNumerically("<", time.Millisecond*10))
		})

		It("should decrypt data efficiently", func() {
			experiment := gmeasure.NewExperiment("Decryption Performance")
			AddReportEntry(experiment.Name, experiment)

			experiment.Sample(func(idx int) {
				experiment.MeasureDuration("decryption", func() {
					_, err := encryptor.Decrypt(ciphertext)
					Expect(err).NotTo(HaveOccurred())
				})
			}, gmeasure.SamplingConfig{N: benchmarkIterations})

			Expect(experiment.GetStats("decryption").DurationFor(gmeasure.StatMean)).To(BeNumerically("<", time.Millisecond*1))
		})
	})
})
