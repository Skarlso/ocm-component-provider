package cmd

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/Skarlso/ocm-component-provider/pkg/ocm"
)

var (
	rootCmd = &cobra.Command{
		Use:   "generate",
		Short: "Create a component version out of the provided helm chart.",
		RunE:  runGenerateCmd,
	}

	rootArgs struct {
		componentName     string
		componentVersion  string
		componentProvider string

		chartName string
		input     string
		output    string
		verbose   bool
	}
)

func init() {
	flag := rootCmd.Flags()
	// Server Configs
	flag.BoolVarP(&rootArgs.verbose, "verbose", "v", false, "--verbose")
	flag.StringVarP(&rootArgs.input, "input", "i", "", "--input {chart.tar.gz|chart/}")
	flag.StringVarP(&rootArgs.output, "output", "o", ".", "--output dir")
	flag.StringVarP(&rootArgs.chartName, "chart-name", "n", "", "--chart-name if it doesn't match the name of the file")
	flag.StringVarP(&rootArgs.componentName, "component", "c", "", "--component github.com/open-component-model/component")
	flag.StringVarP(&rootArgs.componentVersion, "version", "r", "0.1.0", "--version 0.1.0")
	flag.StringVarP(&rootArgs.componentProvider, "provider", "p", "ocm", "--provider ocm")
}

func runGenerateCmd(_ *cobra.Command, _ []string) error {
	comp := &ocm.Component{
		ObjectMeta: v1.ObjectMeta{
			Name:    rootArgs.componentName,
			Version: rootArgs.componentVersion,
			Provider: v1.Provider{
				Name: compdesc.ProviderName(rootArgs.componentProvider),
			},
		},
	}

	if err := ocm.AddResource(comp, ocm.ResourceOptions{Location: rootArgs.input, ChartName: rootArgs.chartName}); err != nil {
		return fmt.Errorf("could not add resource: %w", err)
	}

	components := &ocm.Components{
		Component: []*ocm.Component{comp},
	}

	content, err := yaml.Marshal(components)
	if err != nil {
		return err
	}

	fmt.Println(string(content))

	return nil
}

// Execute runs the main serve command.
func Execute() error {
	return rootCmd.Execute()
}
