/*
Copyright © 2025 Tailor Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/k1LoW/duration"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
	"github.com/tailor-platform/patterner/config"
	"github.com/tailor-platform/patterner/tailor"
)

var (
	outOctocovPath         string
	withLintWarnings       bool
	withCoverageFullReport bool
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "retrieve and display metrics about the resources in the workspace",
	Long:  `retrieve and display metrics about the resources in the specified workspace.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		spi.Start()
		defer spi.Stop()
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if workspaceID != "" {
			cfg.WorkspaceID = workspaceID
		}
		c, err := tailor.New(cfg)
		if err != nil {
			return err
		}
		d, err := duration.Parse(since)
		if err != nil {
			return err
		}
		s := time.Now().Add(-d)
		resources, err := c.Resources(cmd.Context(), tailor.WithExecutionResults(&s))
		if err != nil {
			return err
		}
		spi.Disable()

		if withLintWarnings {
			fmt.Println("Lint warnings")
			fmt.Println("============================================================")
			warns, err := c.Lint(resources)
			if err != nil {
				return err
			}
			for _, w := range warns {
				fmt.Printf("[%s] %s: %s\n", w.Type, w.Name, w.Message)
			}
			fmt.Println()
		}

		if withCoverageFullReport {
			fmt.Println("Coverage")
			fmt.Println("============================================================")
			coverages, err := c.Coverage(resources)
			if err != nil {
				return err
			}
			for _, rc := range coverages {
				var cover float64
				if rc.TotalSteps > 0 {
					cover = float64(float64(rc.CoveredSteps)/float64(rc.TotalSteps)) * 100
				}
				fmt.Printf("%5s%% [%d/%d] %s\n", fmt.Sprintf("%.1f", cover), rc.CoveredSteps, rc.TotalSteps, rc.Name)
			}
			fmt.Println()
		}

		if withCoverageFullReport || withLintWarnings {
			fmt.Println("Metrics")
			fmt.Println("============================================================")
		}

		metrics, err := c.Metrics(resources)
		if err != nil {
			return err
		}
		table := tablewriter.NewTable(os.Stdout,
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
				Borders: tw.BorderNone,
				Symbols: tw.NewSymbols(tw.StyleNone),
				Settings: tw.Settings{
					Lines: tw.Lines{
						ShowTop:        tw.Off,
						ShowBottom:     tw.Off,
						ShowHeaderLine: tw.Off,
						ShowFooterLine: tw.Off,
					},
					Separators: tw.Separators{
						ShowHeader:     tw.Off,
						ShowFooter:     tw.Off,
						BetweenRows:    tw.Off,
						BetweenColumns: tw.Off,
					},
				},
			})),
			tablewriter.WithHeaderConfig(tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoFormat: tw.Off,
					Alignment:  tw.AlignLeft,
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: tw.Space, Right: tw.Space, Top: tw.Empty, Bottom: tw.Empty},
				},
			}),
			tablewriter.WithRowConfig(tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoFormat: tw.Off,
				},
				ColumnAligns: []tw.Align{tw.AlignLeft, tw.AlignRight},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: tw.Space, Right: tw.Space, Top: tw.Empty, Bottom: tw.Empty},
				},
			}),
		)
		data := make([][]string, 0, len(metrics))
		for _, m := range metrics {
			if m.Error != nil {
				data = append(data, []string{m.Name, fmt.Sprintf("Error: %v", m.Error)})
				continue
			}
			if m.Value == (math.Round(m.Value*10) / 10) {
				data = append(data, []string{m.Name, fmt.Sprintf("%.0f%s", m.Value, m.Unit)})
			} else {
				data = append(data, []string{m.Name, fmt.Sprintf("%.1f%s", m.Value, m.Unit)})
			}
		}
		if err := table.Bulk(data); err != nil {
			return err
		}
		if err := table.Render(); err != nil {
			return err
		}
		if outOctocovPath != "" {
			metricSet := &CustomMetricSet{
				Key:  "workspace_metrics",
				Name: "Workspace metrics using [Patterner](https://github.com/tailor-platform/patterner)",
				Metadata: []*MetadataKV{
					{
						Key:   "workspace_id",
						Value: cfg.WorkspaceID,
					},
				},
			}
			for _, m := range metrics {
				metricSet.Metrics = append(metricSet.Metrics, &CustomMetric{
					Key:   m.Key,
					Name:  m.Name,
					Value: m.Value,
					Unit:  m.Unit,
				})
			}
			metricSet.Acceptables = append(metricSet.Acceptables, cfg.Metrics.Octocov.Acceptables...)

			csets := []*CustomMetricSet{metricSet}
			b, err := json.MarshalIndent(csets, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(outOctocovPath, b, 0644); err != nil { //nolint:gosec
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(metricsCmd)
	metricsCmd.Flags().StringVarP(&since, "since", "s", "30min", "only consider executions since the given duration (e.g., 24hours, 30min, 15sec)")
	metricsCmd.Flags().StringVarP(&outOctocovPath, "out-octocov-path", "", "", "output the metrics in octocov custom metrics format to the specified file (e.g., ./metrics.json)")
	metricsCmd.Flags().BoolVarP(&withLintWarnings, "with-lint-warnings", "", false, "display the lint warnings along with the metrics")
	metricsCmd.Flags().BoolVarP(&withCoverageFullReport, "with-coverage-full-report", "", false, "display the coverage full report along with the metrics")
}

// copy from github.com/k1LoW/octocov/report
// because octocov use tablewriter v0.
type MetadataKV struct {
	Key   string `json:"key"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value"`
}

type CustomMetricSet struct {
	Key         string          `json:"key"`
	Name        string          `json:"name,omitempty"`
	Metadata    []*MetadataKV   `json:"metadata,omitempty"`
	Metrics     []*CustomMetric `json:"metrics"`
	Acceptables []string        `json:"acceptables,omitempty"`
}

type CustomMetric struct {
	Key   string  `json:"key"`
	Name  string  `json:"name,omitempty"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit,omitempty"`
}
