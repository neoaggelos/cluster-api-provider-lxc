package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lxc/cluster-api-provider-incus/internal/exp/simplestreams/index"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

func newShowCmd() *cobra.Command {
	var flags struct {
		rootDir string

		output string

		product string
		os      string
		release string
		arch    string
		itype   string

		incus bool
		lxd   bool
	}

	cmd := &cobra.Command{
		Use:     "show",
		Short:   "Show images in a simplestreams index",
		GroupID: "operations",

		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch flags.output {
			case "images-json", "index-json", "pretty":
			default:
				return fmt.Errorf("invalid argument value --output=%q. Must be one of [pretty, images-json, index-json]", flags.output)
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			index, err := index.GetOrCreateIndex(flags.rootDir)
			if err != nil {
				return fmt.Errorf("failed to read simplestreams index: %w", err)
			}

			if flags.output == "images-json" {
				b, err := json.MarshalIndent(index.Products, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}

				fmt.Println(string(b))
				return nil
			}

			if flags.output == "index-json" {
				b, err := json.MarshalIndent(index.Index, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}

				fmt.Println(string(b))
				return nil
			}

			if flags.output == "pretty" {
				fmt.Println("| NAME                           | SERIAL       | TYPE            | SRV   | ARCH  |  SIZE   | PATH")
				fmt.Println("|--------------------------------|--------------|-----------------|-------|-------|---------|--------------------------------------------------------")
				for _, productName := range index.Index.Index["images"].Products {
					product := index.Products.Products[productName]
					switch {
					case flags.product != "" && flags.product != productName:
						continue
					case flags.arch != "" && product.Architecture != flags.arch:
						continue
					case flags.os != "" && product.OperatingSystem != flags.os:
						continue
					case flags.release != "" && product.Release != flags.release:
						continue
					}

					for versionName, version := range product.Versions {
						for _, item := range version.Items {
							switch item.FileType {
							case "incus_combined.tar.gz":
								if !flags.incus && flags.lxd || flags.itype == lxc.VirtualMachine {
									continue
								}
								fmt.Printf("| %-30s | %s | container       | incus | %s | %4v MB | %s\n", productName, versionName, product.Architecture, item.Size/1024/1024, item.Path)
							case "lxd_combined.tar.gz":
								if !flags.lxd && flags.incus || flags.itype == lxc.VirtualMachine {
									continue
								}
								fmt.Printf("| %-30s | %s | container       | lxd   | %s | %4v MB | %s\n", productName, versionName, product.Architecture, item.Size/1024/1024, item.Path)
							case "disk-kvm.img":
								if !flags.incus && flags.lxd || flags.itype == lxc.Container {
									continue
								}
								if _, ok := version.Items["incus.tar.xz"]; !ok {
									continue
								}
								fmt.Printf("| %-30s | %s | virtual-machine | incus | %s | %4v MB | %s\n", productName, versionName, product.Architecture, item.Size/1024/1024, item.Path)
							case "disk1.img":
								if !flags.lxd && flags.incus || flags.itype == lxc.Container {
									continue
								}
								if _, ok := version.Items["lxd.tar.xz"]; !ok {
									continue
								}
								fmt.Printf("| %-30s | %s | virtual-machine | lxd   | %s | %4v MB | %s\n", productName, versionName, product.Architecture, item.Size/1024/1024, item.Path)
							}
						}
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.rootDir, "root-dir", "",
		"Simplestreams index directory")
	cmd.Flags().StringVar(&flags.output, "output", "pretty",
		"Output format. Must be one of [pretty]")
	cmd.Flags().StringVar(&flags.product, "product", "",
		"Filter available products by name")
	cmd.Flags().StringVar(&flags.os, "os", "",
		"Filter available products by operating system")
	cmd.Flags().StringVar(&flags.arch, "arch", "",
		"Filter available products by architecture")
	cmd.Flags().StringVar(&flags.release, "release", "",
		"Filter available products by release name")
	cmd.Flags().StringVar(&flags.itype, "type", "",
		"Filter available products by image type. Must be one of [container, virtual-machine]")
	cmd.Flags().BoolVar(&flags.incus, "incus", false,
		"Filter available products for Incus")
	cmd.Flags().BoolVar(&flags.lxd, "lxd", false,
		"Filter available products for LXD")

	return cmd
}
