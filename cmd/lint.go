/*
Copyright © 2025 Ken'ichiro Oyama <k1lowxb@gmail.com>

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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tailor-platform/patterner/config"
	"github.com/tailor-platform/patterner/tailor"
)

// lintCmd represents the lint command
var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "lint the resources in the workspace",
	Long:  `lint the resources in the specified workspace.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		warns, err := c.Lint(resources)
		if err != nil {
			return err
		}
		for _, w := range warns {
			fmt.Printf("[%s] %s: %s\n", w.Type, w.Name, w.Message)
		}
		if len(warns) > 0 {
			return errors.New("lint warnings found")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
}
