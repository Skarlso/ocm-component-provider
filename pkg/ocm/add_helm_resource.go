package ocm

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containers/storage/pkg/archive"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/runtime"
	"gopkg.in/yaml.v3"
)

// HelmInput defines a helm type input.
type HelmInput struct {
	Version string `json:"version"`
	Type    string `json:"type"`
	Path    string `json:"path"`
}

func (h HelmInput) ToUnstructured() (*runtime.UnstructuredTypedObject, error) {
	return &runtime.UnstructuredTypedObject{
		Object: map[string]interface{}{
			getTagName(h, h.Version): h.Version,
			getTagName(h, h.Type):    h.Type,
			getTagName(h, h.Path):    h.Path,
		},
	}, nil
}

// ResourceOptions provides options to override resources with.
type ResourceOptions struct {
	ChartName string
	Location  string
}

// AddHelmResource adds a resource to a component descriptor based on the given Helm Chart details.
func AddHelmResource(cd *Component, opts ResourceOptions) error {
	info, err := os.Stat(opts.Location)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return addResourceFromFolder(cd, opts)
	}

	return addResourceFromFile(cd, opts)
}

type ChartFile struct {
	AppVersion string `yaml:"appVersion"`
	Version    string `yaml:"version"`
	Name       string `yaml:"name"`
}

func addResourceFromFile(cd *Component, opts ResourceOptions) error {
	name := opts.ChartName
	if name == "" {
		base := filepath.Base(opts.Location)
		name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	chartFile, err := getChartFileFromTar(name, opts.Location)
	if err != nil {
		return fmt.Errorf("error getting chart file version and name: %v", err)
	}

	input := HelmInput{
		Version: chartFile.Version,
		Type:    "helm",
		Path:    opts.Location,
	}

	cd.Resources = append(cd.Resources, Resource{
		ResourceMeta: compdesc.ResourceMeta{
			ElementMeta: compdesc.ElementMeta{
				Name:    chartFile.Name,
				Version: chartFile.Version,
			},
			Type:     resourcetypes.HELM_CHART,
			Relation: compdesc.LocalRelation,
		},
		Input: input,
	})

	return nil
}

func getChartFileFromTar(name, location string) (*ChartFile, error) {
	temp, err := os.MkdirTemp("", "tar")
	if err != nil {
		return nil, fmt.Errorf("could not create temp directory: %w", err)
	}

	defer os.RemoveAll(temp)

	content, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	if err := archive.Untar(bytes.NewBuffer(content), temp, nil); err != nil {
		return nil, fmt.Errorf("error untaring file: %w", err)
	}

	chartFileContent, err := os.ReadFile(filepath.Join(temp, name, "Chart.yaml"))
	if err != nil {
		return nil, fmt.Errorf("error reading Chart.yaml file: %w", err)
	}

	return parseChartFileContent(chartFileContent)
}

func parseChartFileContent(content []byte) (*ChartFile, error) {
	chartFile := &ChartFile{}
	if err := yaml.Unmarshal(content, chartFile); err != nil {
		return nil, fmt.Errorf("error unmarshalling file: %w", err)
	}

	version := chartFile.Version
	if version == "" {
		version = chartFile.AppVersion
		if version == "" {
			return nil, fmt.Errorf("could not determine chart version")
		}
	}

	return chartFile, nil
}

func addResourceFromFolder(cd *Component, opts ResourceOptions) error {
	chartFileContent, err := os.ReadFile(filepath.Join(opts.Location, "Chart.yaml"))
	if err != nil {
		return fmt.Errorf("error reading Chart.yaml file: %w", err)
	}

	chartFile, err := parseChartFileContent(chartFileContent)
	if err != nil {
		return fmt.Errorf("error parsing Chart.yaml file: %w", err)
	}

	input := HelmInput{
		Version: chartFile.Version,
		Type:    "helm",
		Path:    opts.Location,
	}

	cd.Resources = append(cd.Resources, Resource{
		ResourceMeta: compdesc.ResourceMeta{
			ElementMeta: compdesc.ElementMeta{
				Name:    chartFile.Name,
				Version: chartFile.Version,
			},
			Type:     resourcetypes.HELM_CHART,
			Relation: compdesc.LocalRelation,
		},
		Input: input,
	})

	return nil
}
