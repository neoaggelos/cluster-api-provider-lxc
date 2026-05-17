package lxc

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

var (
	wellKnownImages = map[string]func(string) ImageFamily{
		"ubuntu":   UbuntuImage,
		"debian":   DebianImage,
		"images":   DefaultImage,
		"capi":     func(in string) ImageFamily { return CapnImage(in) },
		"capi-stg": func(in string) ImageFamily { return CapnStagingImage(in) },
		"capi-dev": func(in string) ImageFamily { return CapnDevelopmentImage(in) },
		"kind":     func(in string) ImageFamily { return KindestNodeImage(in) },
	}
)

func ParseImage(imageName string) (ImageFamily, bool, error) {
	parts := strings.Split(imageName, ":")
	if len(parts) != 2 {
		return nil, false, nil
	}
	if f, ok := wellKnownImages[parts[0]]; ok {
		return f(parts[1]), true, nil
	}

	return nil, false, fmt.Errorf("unknown image prefix %q, must be one of %v", parts[0], slices.Collect(maps.Keys(wellKnownImages)))
}
