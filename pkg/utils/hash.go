package utils

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"path/filepath"
	"sort"

	helm "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

// Hash creates a 64-bit FNV-1 hash of the Helm chart including the values
func Hash(chart *helm.Chart, values chartutil.Values) string {
	algorithm := fnv.New64()

	// copy and sort the helm chart's files
	files := make([]*helm.File, len(chart.Raw))
	copy(files, chart.Raw)
	sort.Slice(files, func(i, j int) bool {
		return files[i] == nil || files[j] == nil || filepath.Base(files[i].Name) > filepath.Base(files[j].Name)
	})

	// add files to the hash
	for _, file := range files {
		if file != nil {
			_, _ = algorithm.Write([]byte(filepath.Base(file.Name)))
			_, _ = algorithm.Write(file.Data)
		}
	}

	// add values encoded as JSON to the hash
	if data, err := json.Marshal(values); err == nil {
		_, _ = algorithm.Write(data)
	}

	// convert hash sum to hex string
	return fmt.Sprintf("%016x", algorithm.Sum64())
}
