package utils

import (
	"encoding/json"
	"fmt"
	"hash/fnv"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

// Hash creates a 32-bit FNV-1 hash of the Helm chart including the values
func Hash(chart *chart.Chart, values chartutil.Values) string {
	algorithm := fnv.New64()

	for _, file := range chart.Files {
		if file != nil {
			_, _ = algorithm.Write(file.Data)
		}
	}

	if data, err := json.Marshal(values); err == nil {
		_, _ = algorithm.Write(data)
	}

	return fmt.Sprintf("%08x", algorithm.Sum64())
}
