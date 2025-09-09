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

	"github.com/spf13/cobra"
	"github.com/tailor-platform/patterner/config"
	"github.com/tailor-platform/patterner/tailor"
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
		resources, err := c.Resources(cmd.Context())
		if err != nil {
			return err
		}
		spi.Disable()
		metrics, err := c.Metrics(resources)
		if err != nil {
			return err
		}
		b, err := json.MarshalIndent(metrics, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(metricsCmd)
}
