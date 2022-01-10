/*
	Copyright NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package cmd

import (
	_ "embed"
	cmdhelper "github.com/openziti/ziti/ziti/cmd/ziti/cmd/helpers"
	"github.com/openziti/ziti/ziti/cmd/ziti/cmd/templates"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	optionWSS          = "wss"
	defaultWSS         = false
	wssDescription     = "Create an edge router config with wss enabled"
	optionPrivate      = "private"
	defaultPrivate     = false
	privateDescription = "Create a private router config"
)

var (
	createConfigRouterEdgeLong = templates.LongDesc(`
		Creates the edge router config
`)

	createConfigRouterEdgeExample = templates.Examples(`
		# Create the edge router config for a router named my_router
		ziti create config router edge --routerName my_router
	`)
)

// CreateConfigRouterEdgeOptions the options for the edge command
type CreateConfigRouterEdgeOptions struct {
	CreateConfigRouterOptions

	WssEnabled bool
	IsPrivate  bool
}

//go:embed config_templates/edge.router.yml
var routerConfigEdgeTemplate string

// NewCmdCreateConfigRouterEdge creates a command object for the "edge" command
func NewCmdCreateConfigRouterEdge(data *ConfigTemplateValues) *cobra.Command {
	options := &CreateConfigRouterEdgeOptions{}

	cmd := &cobra.Command{
		Use:     "edge",
		Short:   "Create an edge router config",
		Aliases: []string{"edge"},
		Long:    createConfigRouterEdgeLong,
		Example: createConfigRouterEdgeExample,
		PreRun: func(cmd *cobra.Command, args []string) {
			// Setup logging
			var logOut *os.File
			if options.Verbose {
				logrus.SetLevel(logrus.DebugLevel)
				// Only print log to stdout if not printing config to stdout
				if strings.ToLower(options.Output) != "stdout" {
					logOut = os.Stdout
				} else {
					logOut = os.Stderr
				}
				logrus.SetOutput(logOut)
			}

			// Update edge router specific values with options passed in
			data.EdgeRouterName = options.RouterName
			data.WssEnabled = options.WssEnabled
			data.IsPrivate = options.IsPrivate
		},
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.run(data)
			cmdhelper.CheckErr(err)
		},
	}

	options.addCreateFlags(cmd)
	options.addFlags(cmd)

	return cmd
}

func (options *CreateConfigRouterEdgeOptions) addFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&options.WssEnabled, optionWSS, defaultWSS, wssDescription)
	cmd.Flags().BoolVar(&options.IsPrivate, optionPrivate, defaultPrivate, privateDescription)
	cmd.PersistentFlags().StringVarP(&options.RouterName, optionRouterName, "n", "", "name of the router")
	err := cmd.MarkPersistentFlagRequired(optionRouterName)
	if err != nil {
		return
	}
}

// run implements the command
func (options *CreateConfigRouterEdgeOptions) run(data *ConfigTemplateValues) error {
	// Ensure private and wss are not both used
	if options.IsPrivate && options.WssEnabled {
		logrus.Fatal("Flags for private and wss configs are mutually exclusive. You must choose private or wss, not both")
		return errors.New("Flags for private and wss configs are mutually exclusive.")
	}

	tmpl, err := template.New("router-config").Parse(routerConfigEdgeTemplate)
	if err != nil {
		return err
	}

	var f *os.File
	if strings.ToLower(options.Output) != "stdout" {
		// Check if the path exists, fail if it doesn't
		basePath := filepath.Dir(options.Output) + "/"
		if _, err := os.Stat(filepath.Dir(basePath)); os.IsNotExist(err) {
			logrus.Fatalf("Provided path: [%s] does not exist\n", basePath)
			return err
		}

		f, err = os.Create(options.Output)
		logrus.Debugf("Created output file: %s", options.Output)
		if err != nil {
			return errors.Wrapf(err, "unable to create config file: %s", options.Output)
		}
	} else {
		f = os.Stdout
	}
	defer func() { _ = f.Close() }()

	if err := tmpl.Execute(f, data); err != nil {
		return errors.Wrap(err, "unable to execute template")
	}

	logrus.Debugf("Edge Router configuration generated successfully and written to: %s", options.Output)

	return nil
}
