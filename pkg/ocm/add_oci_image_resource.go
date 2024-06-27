package ocm

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/containers/storage/pkg/archive"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// OCIImageAccess defines an oci image access resource.
type OCIImageAccess struct {
	Type           string `json:"type"`
	ImageReference string `json:"imageReference"`
}

func (o OCIImageAccess) ToUnstructured() (*runtime.UnstructuredTypedObject, error) {
	return &runtime.UnstructuredTypedObject{
		Object: map[string]interface{}{
			getTagName(o, o.Type):           o.Type,
			getTagName(o, o.ImageReference): o.ImageReference,
		},
	}, nil
}

// AddImageResource adds an oci image resource to a component.
func AddImageResource(cd *Component, opts ResourceOptions) error {
	images, err := scanFolderOrArchive(opts)
	if err != nil {
		return fmt.Errorf("could not scan folder or archive: %w", err)
	}

	for i, image := range images {
		access := OCIImageAccess{
			Type:           resourcetypes.OCI_IMAGE,
			ImageReference: fmt.Sprintf("# replace reference for image value: %s in file %s", image.image, image.file),
		}

		cd.Resources = append(cd.Resources, Resource{
			ResourceMeta: compdesc.ResourceMeta{
				ElementMeta: compdesc.ElementMeta{
					Name:    "image" + strconv.Itoa(i),
					Version: "0.0.1",
				},
				Type: resourcetypes.OCI_ARTIFACT,
			},
			Access: access,
		})
	}

	return nil
}

type replace struct {
	file  string
	image string
}

func scanFolderOrArchive(opts ResourceOptions) ([]replace, error) {
	info, err := os.Stat(opts.Location)
	if err != nil {
		return nil, fmt.Errorf("could not find folder %s: %w", opts.Location, err)
	}

	if info.IsDir() {
		return scan(opts, false)
	}

	return scanArchive(opts)
}

func scanArchive(opts ResourceOptions) ([]replace, error) {
	temp, err := os.MkdirTemp("", "tar")
	if err != nil {
		return nil, fmt.Errorf("could not create temp directory: %w", err)
	}

	content, err := os.ReadFile(opts.Location)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	if err := archive.Untar(bytes.NewBuffer(content), temp, nil); err != nil {
		return nil, fmt.Errorf("error untaring file: %w", err)
	}

	opts.Location = temp

	return scan(opts, true)
}

func scan(opts ResourceOptions, cleanup bool) ([]replace, error) {
	if cleanup {
		defer os.RemoveAll(opts.Location)
	}

	templates := filepath.Join(opts.Location, opts.ChartName, "templates")
	if _, err := os.Stat(templates); err != nil {
		return nil, fmt.Errorf("failed to check for templates under %s: %w", templates, err)
	}

	var images []replace
	err := filepath.Walk(templates, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// skip test folders
		if strings.Contains(path, "tests") {
			return nil
		}

		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file: %w", err)
			}

			// Because this is a helm template file, we can't parse it
			// into a map. Hence, our old friend, Contains.
			for _, l := range bytes.Split(content, []byte("\n")) {
				index := bytes.Index(l, []byte("image:"))
				if index == -1 {
					continue
				}

				elems := string(l[index+len("image:"):])
				images = append(images, replace{
					file:  filepath.Base(path),
					image: elems,
				})
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error scanning templates: %w", err)
	}

	return images, nil
}
