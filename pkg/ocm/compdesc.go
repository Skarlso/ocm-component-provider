package ocm

import (
	"fmt"
	"reflect"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Components struct {
	Component []*Component `json:"components"`
}

// Resource represents an input resource.
type Resource struct {
	ocm.ResourceMeta `json:",inline"`
	Input            runtime.UnstructuredConverter `json:"input,omitempty"`
	Access           runtime.UnstructuredConverter `json:"access,omitempty"`
}

type Component struct {
	v1.ObjectMeta `json:",inline"`
	Resources     []Resource `json:"resources"`
}

// getTagName returns the json value for the struct. Used so we don't hardcode map values for
// unstructured struct types.
func getTagName(v interface{}, fieldName string) string {
	t := reflect.TypeOf(v)
	field, found := t.FieldByName(fieldName)
	if !found {
		panic(fmt.Errorf("field '%s' not found in struct", fieldName))
	}
	return field.Tag.Get("json")
}
