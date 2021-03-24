package utils

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sort"

	helm "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

// Hash creates a 32-bit FNV-1 hash of the Helm chart including the values
func Hash(chart *helm.Chart, values chartutil.Values) string {
	algorithm := fnv.New64()

	// copy and sort the helm chart's files
	files := make([]*helm.File, len(chart.Files))
	copy(files, chart.Files)
	sort.Slice(files, func(i, j int) bool {
		return files[i] == nil || files[j] == nil || files[i].Name > files[j].Name
	})

	// add files to the hash
	for _, file := range files {
		if file != nil {
			_, _ = algorithm.Write([]byte(file.Name))
			_, _ = algorithm.Write(file.Data)
		}
	}

	// add values encoded as JSON to the hash
	if data, err := json.Marshal(values); err == nil {
		_, _ = algorithm.Write(data)
	}

	// convert hash sum to hex string
	return fmt.Sprintf("%08x", algorithm.Sum64())
}
