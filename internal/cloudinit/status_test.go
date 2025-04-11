package cloudinit_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/lxc/cluster-api-provider-incus/internal/cloudinit"

	. "github.com/onsi/gomega"
)

func TestParseStatus(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		g := NewWithT(t)

		status, err := cloudinit.ParseStatus(nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).To(Equal(cloudinit.StatusUnknown))
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		g := NewWithT(t)

		status, err := cloudinit.ParseStatus(bytes.NewReader([]byte(`invalid json`)))
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).To(Equal(cloudinit.StatusUnknown))
	})

	t.Run("Valid", func(t *testing.T) {
		for _, tc := range []struct {
			file         string
			expectStatus cloudinit.Status
		}{
			{file: "testdata/done.json", expectStatus: cloudinit.StatusDone},
			{file: "testdata/running.json", expectStatus: cloudinit.StatusRunning},
			{file: "testdata/error.json", expectStatus: cloudinit.StatusError},
		} {
			t.Run(string(tc.expectStatus), func(t *testing.T) {
				g := NewWithT(t)

				f, err := os.Open(tc.file)
				g.Expect(err).NotTo(HaveOccurred())

				status, err := cloudinit.ParseStatus(f)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(status).To(Equal(tc.expectStatus))
			})
		}
	})
}
