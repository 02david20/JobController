//go:build tools
// +build tools

//go:generate go run github.com/losisin/helm-values-schema-json/v2 -f ../dist/chart/values.yaml -o ../dist/chart/values.schema.json

package hack

import (
	_ "github.com/losisin/helm-values-schema-json/v2"
)
