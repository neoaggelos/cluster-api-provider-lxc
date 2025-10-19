package kini

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

func serializeRuntimeObject(out io.Writer, gv schema.GroupVersion, obj runtime.Object) error {
	info, ok := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
	if !ok {
		return fmt.Errorf("failed to initialize serializer")
	}

	return scheme.Codecs.EncoderForVersion(info.Serializer, gv).Encode(obj, out)
}

func newKiniSetupGenerateSecretCmd() *cobra.Command {
	var flags struct {
		configFile string
		remoteName string

		namespace string
	}

	cmd := &cobra.Command{
		Use:           "generate-secret NAME",
		Short:         "Generate a Kubernetes secret with CAPN credentials from local configuration",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, _, err := lxc.ConfigurationFromLocal(flags.configFile, flags.remoteName, true)
			if err != nil {
				return fmt.Errorf("failed to read local configuration: %w", err)
			}

			if err := serializeRuntimeObject(os.Stdout, corev1.SchemeGroupVersion, opts.ToSecret(args[0], flags.namespace)); err != nil {
				return fmt.Errorf("failed to encode secret: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&flags.namespace, "namespace", "",
		"Set namespace on the generated secret")
	cmd.Flags().StringVar(&flags.configFile, "config-file", "",
		"Read client configuration from file")
	cmd.Flags().StringVar(&flags.remoteName, "remote-name", "",
		"Override remote to use from configuration file")

	return cmd
}
