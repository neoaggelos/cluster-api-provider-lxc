//go:build tools

package tools

import (
	_ "github.com/ahmetb/gen-crd-api-reference-docs"
	_ "sigs.k8s.io/cluster-api/hack/tools/mdbook/embed"
	_ "sigs.k8s.io/cluster-api/hack/tools/mdbook/releaselink"
	_ "sigs.k8s.io/cluster-api/hack/tools/mdbook/tabulate"
)
