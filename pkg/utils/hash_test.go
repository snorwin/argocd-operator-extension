package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/snorwin/argocd-operator-extension/pkg/utils"
)

var _ = Describe("Hash", func() {
	Context("Hash", func() {
		It("should_be_idempotent", func() {
			chart := &chart.Chart{
				Files: []*chart.File{
					{
						Name: "file1.yaml",
						Data: []byte{1, 1, 1, 1},
					},
					{
						Name: "file2.yaml",
						Data: []byte{2, 2, 2, 2},
					},
				},
			}
			values := chartutil.Values{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}

			Ω(utils.Hash(chart, values)).Should(Equal(utils.Hash(chart, values)))
		})
		It("should_be_different_if_values_are_different", func() {
			Ω(utils.Hash(&chart.Chart{}, chartutil.Values{"key1": "value1", "key2": "value2", "key3": "value3"})).
				ShouldNot(Equal(utils.Hash(&chart.Chart{}, chartutil.Values{"key1": "value1", "key2": "value2"})))

			Ω(utils.Hash(&chart.Chart{}, chartutil.Values{"key1": "value1", "key2": "value2", "key3": "value3"})).
				ShouldNot(Equal(utils.Hash(&chart.Chart{}, chartutil.Values{"key1": "value1", "key2": "value2", "key3": "changed"})))
		})
		It("should_be_different_if_chart_files_are_different", func() {
			Ω(utils.Hash(&chart.Chart{Files: []*chart.File{{Name: "file1.yaml", Data: []byte{1, 1, 1, 1}}}}, chartutil.Values{})).
				ShouldNot(Equal(utils.Hash(&chart.Chart{Files: []*chart.File{{Name: "file1.yaml", Data: []byte{1, 1, 1}}}}, chartutil.Values{})))

			Ω(utils.Hash(&chart.Chart{Files: []*chart.File{{Name: "file1.yaml", Data: []byte{1, 1, 1, 1}}}}, chartutil.Values{})).
				ShouldNot(Equal(utils.Hash(&chart.Chart{Files: []*chart.File{{Name: "file2.yaml", Data: []byte{1, 1, 1, 1}}}}, chartutil.Values{})))
		})
		It("should_ignore_file_order_and_nil", func() {
			chart1 := &chart.Chart{
				Files: []*chart.File{
					{
						Name: "file1.yaml",
						Data: []byte{1, 1, 1, 1},
					},
					{
						Name: "file2.yaml",
						Data: []byte{2, 2, 2, 2},
					},
					nil,
				},
			}

			chart2 := &chart.Chart{
				Files: []*chart.File{
					{
						Name: "file2.yaml",
						Data: []byte{2, 2, 2, 2},
					},
					nil,
					{
						Name: "file1.yaml",
						Data: []byte{1, 1, 1, 1},
					},
					nil,
				},
			}

			Ω(utils.Hash(chart1, chartutil.Values{})).Should(Equal(utils.Hash(chart2, chartutil.Values{})))
		})
		It("should_ignore_values_order", func() {
			values1 := chartutil.Values{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}

			values2 := chartutil.Values{
				"key3": "value3",
				"key2": "value2",
				"key1": "value1",
			}

			Ω(utils.Hash(&chart.Chart{}, values1)).Should(Equal(utils.Hash(&chart.Chart{}, values2)))
		})
		It("should_not_modify_file_order", func() {
			expected := &chart.Chart{
				Files: []*chart.File{
					{
						Name: "zzzz",
						Data: []byte{1, 1, 1, 1},
					},
					{
						Name: "xxxxx",
						Data: []byte{2, 2, 2, 2},
					},
					{
						Name: "yyyyy",
						Data: []byte{3, 3, 3, 3},
					},
				},
			}
			actual := &chart.Chart{
				Files: []*chart.File{
					{
						Name: "zzzz",
						Data: []byte{1, 1, 1, 1},
					},
					{
						Name: "xxxxx",
						Data: []byte{2, 2, 2, 2},
					},
					{
						Name: "yyyyy",
						Data: []byte{3, 3, 3, 3},
					},
				},
			}

			utils.Hash(actual, chartutil.Values{})

			Ω(actual.Files).Should(Equal(expected.Files))
		})
	})
})
