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
	"fmt"
	"time"

	"github.com/k1LoW/duration"
	"github.com/spf13/cobra"
	"github.com/tailor-platform/patterner/config"
	"github.com/tailor-platform/patterner/tailor"
)

var (
	since      string
	fullReport bool
)

var coverageCmd = &cobra.Command{
	Use:   "coverage",
	Short: "display the pipeline resolver step coverage",
	Long:  `display the pipeline resolver step coverage.`,
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
		opts := []tailor.ResourceOption{
			tailor.WithExecutionResults(&s),
			tailor.WithoutApplications(),
			tailor.WithoutStateFlow(),
			tailor.WithoutTailorDB(),
		}
		resources, err := c.Resources(cmd.Context(), opts...)
		if err != nil {
			return err
		}
		spi.Disable()
		coverage, err := c.Coverage(resources)
		if err != nil {
			return err
		}
		var total, covered int
		for _, rc := range coverage {
			if fullReport {
				cover := float64(float64(rc.CoveredSteps)/float64(rc.TotalSteps)) * 100
				fmt.Printf("%5s%% [%d/%d] %s\n", fmt.Sprintf("%.1f", cover), rc.CoveredSteps, rc.TotalSteps, rc.Name)
			}
			total += rc.TotalSteps
			covered += rc.CoveredSteps
		}
		if fullReport {
			fmt.Println()
		}
		fmt.Printf("%s %.1f%% [%d/%d]\n", "Pipeline Resolver Step Coverage", float64(float64(covered)/float64(total))*100, covered, total)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(coverageCmd)
	coverageCmd.Flags().StringVarP(&since, "since", "s", "30min", "only consider executions since the given duration (e.g., 24hours, 30min, 15sec)")
	coverageCmd.Flags().BoolVarP(&fullReport, "full-report", "f", false, "display full report")
}
