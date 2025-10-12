package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lxc/incus/v6/shared/simplestreams"
)

// Index manages a simplestreams index from a local directory
type Index struct {
	Index    simplestreams.Stream
	Products simplestreams.Products

	rootDir string
}

// GetOrCreateIndex opens a simplestreams index given a local root directory.
func GetOrCreateIndex(rootDir string) (*Index, error) {
	if rootDir == "" {
		if dir, err := os.Getwd(); err != nil {
			return nil, fmt.Errorf("failed to retrieve current directory: %w", err)
		} else {
			rootDir = dir
		}
	}

	if err := os.MkdirAll(filepath.Join(rootDir, "streams", "v1"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create streams/v1 directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(rootDir, "images"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create images directory: %w", err)
	}

	// parse index
	var index simplestreams.Stream
	if indexJSON, err := os.ReadFile(filepath.Join(rootDir, "streams", "v1", "index.json")); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read streams/v1/index.json: %w", err)
		}

		// initialize new index
		index = simplestreams.Stream{
			Format: "index:1.0",
			Index: map[string]simplestreams.StreamIndex{
				"images": {
					DataType: "image-downloads",
					Path:     "streams/v1/images.json",
					Format:   "products:1.0",
				},
			},
		}
	} else if err := json.Unmarshal(indexJSON, &index); err != nil {
		return nil, fmt.Errorf("failed to parse streams/v1/index.json: %w", err)
	}

	// parse products
	var products simplestreams.Products
	if productsJSON, err := os.ReadFile(filepath.Join(rootDir, "streams", "v1", "images.json")); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read streams/v1/images.json: %w", err)
		}

		// initialize new product index
		products = simplestreams.Products{
			ContentID: "images",
			DataType:  "image-downloads",
			Format:    "products:1.0",
			Products:  map[string]simplestreams.Product{},
		}
	} else if err := json.Unmarshal(productsJSON, &products); err != nil {
		return nil, fmt.Errorf("failed to parse streams/v1/images.json: %w", err)
	}

	return &Index{
		rootDir:  rootDir,
		Index:    index,
		Products: products,
	}, nil
}
