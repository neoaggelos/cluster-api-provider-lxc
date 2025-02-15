package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/lxc/incus/v6/shared/simplestreams"
	"github.com/spf13/pflag"
	"k8s.io/component-base/logs"
	logsv1 "k8s.io/component-base/logs/api/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	rootDir   string
	imagesDir string

	imageType string

	metadataFile string
	rootfsFile   string

	setupLog   = klog.Background().WithName("setup")
	logOptions = logs.NewOptions()

	log klog.Logger
	ctx context.Context
)

func init() {
	pflag.StringVar(&rootDir, "root-dir", "", "Specify root directory of simplestreams repository. If not specified, the current directory is used.")
	pflag.StringVar(&imagesDir, "images-dir", "images", "Specify directory to store images. This is a path relative to --root-dir.")

	pflag.StringVar(&imageType, "image-type", "", "One of 'container', 'incus-vm', 'lxd-vm'.")
	pflag.StringVar(&metadataFile, "metadata", "", "Path to metadata tarball (for virtual machines) or unified tarball (for containers).")
	pflag.StringVar(&rootfsFile, "rootfs", "", "Path to rootfs image (for virtual machines).")

	logsv1.AddFlags(logOptions, pflag.CommandLine)
}

func main() {
	pflag.Parse()

	if err := logsv1.ValidateAndApply(logOptions, nil); err != nil {
		setupLog.Error(err, "Unable to start manager")
		os.Exit(1)
	}
	log = klog.Background()
	ctx = klog.NewContext(ctrl.SetupSignalHandler(), log)

	if rootDir == "" {
		log.V(1).Info("--root-dir not specified, using current directory")
		if dir, err := os.Getwd(); err != nil {
			log.Error(err, "Error: failed to check current directory")
			os.Exit(1)
		} else {
			rootDir = dir
		}
	}

	log = log.WithValues("rootDir", rootDir, "imagesDir", imagesDir)

	log.Info("Configuring simplestreams repository")
	if err := os.MkdirAll(filepath.Join(rootDir, "streams", "v1"), 0755); err != nil {
		log.Error(err, "Failed to create index directory")
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Join(rootDir, imagesDir), 0755); err != nil {
		log.Error(err, "Failed to create images directory")
		os.Exit(1)
	}

	// parse index
	var index *simplestreams.Stream
	if indexJSON, err := os.ReadFile(filepath.Join(rootDir, "streams", "v1", "index.json")); err != nil {
		if !os.IsNotExist(err) {
			log.Error(err, "Failed to read streams/v1/index.json")
			os.Exit(1)
		}
		// create new index
		index = &simplestreams.Stream{
			Format: "index:1.0",
			Index: map[string]simplestreams.StreamIndex{
				"images": {
					DataType: "image-downloads",
					Path:     "streams/v1/images.json",
					Format:   "products:1.0",
				},
			},
		}
	} else if err := json.Unmarshal(indexJSON, index); err != nil {
		log.Error(err, "Failed to parse streams/v1/index.json")
		os.Exit(1)
	}

	// parse products
	var products *simplestreams.Products
	if productsJSON, err := os.ReadFile(filepath.Join(rootDir, "streams", "v1", "images.json")); err != nil {
		if !os.IsNotExist(err) {
			log.Error(err, "Failed to read streams/v1/images.json")
			os.Exit(1)
		}
		// create new product index
		products = &simplestreams.Products{
			ContentID: "images",
			DataType:  "image-downloads",
			Format:    "products:1.0",
			Products:  map[string]simplestreams.Product{},
		}
	} else if err := json.Unmarshal(productsJSON, products); err != nil {
		log.Error(err, "Failed to parse streams/v1/images.json")
		os.Exit(1)
	}

	switch imageType {
	case "container":
		if metadataFile == "" {
			log.Error(nil, "--metadata argument must be specified for container images")
			os.Exit(1)
		}
		if rootfsFile != "" {
			log.Error(nil, "--rootfs argument is not supported for container images")
			os.Exit(1)
		}

		if err := addUnifiedImageTarball(rootDir, imagesDir, index, products, metadataFile); err != nil {
			log.Error(err, "Failed to add container image to simplestreams index")
			os.Exit(1)
		}
	case "incus-vm":
		log.Error(nil, "--image-type=incus-vm not implemented yet")
	case "lxd-vm":
		log.Error(nil, "--image-type=lxd-vm not implemented yet")
	}

}
