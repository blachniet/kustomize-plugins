package pullpolicy

import (
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/image"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Filter struct {
	Images []imageEntry `yaml:"images" json:"images"`

	// FsSlice contains the FieldSpecs to locate an image field,
	// e.g. Path: "spec/myContainers[]"
	FsSlice types.FsSlice `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

type imageEntry struct {
	Name          string `yaml:"name" json:"name"`
	NewPullPolicy string `yaml:"newPullPolicy" json:"newPullPolicy"`
}

func (f *Filter) Default() error {
	if f.FsSlice == nil || f.FsSlice.Len() == 0 {
		f.FsSlice = []types.FieldSpec{
			{
				Path:               "spec/containers[]",
				CreateIfNotPresent: true,
			},
			{
				Path:               "spec/initContainers[]",
				CreateIfNotPresent: true,
			},
			{
				Path:               "spec/template/spec/containers[]",
				CreateIfNotPresent: true,
			},
			{
				Path:               "spec/template/spec/initContainers[]",
				CreateIfNotPresent: true,
			},
		}
	}

	return nil
}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(f.filter)).Filter(nodes)
	return nodes, err
}

func (f Filter) filter(node *yaml.RNode) (*yaml.RNode, error) {
	// FsSlice is an allowlist, not a denyList, so to deny
	// something via configuration a new config mechanism is
	// needed. Until then, hardcode it.
	if f.isOnDenyList(node) {
		return node, nil
	}
	if err := node.PipeE(fsslice.Filter{
		FsSlice:  f.FsSlice,
		SetValue: f.updateContainerSequence,
	}); err != nil {
		return nil, err
	}
	return node, nil
}

func (f Filter) updateContainerSequence(rn *yaml.RNode) error {
	if err := yaml.ErrorIfInvalid(rn, yaml.SequenceNode); err != nil {
		return err
	}

	return rn.VisitElements(func(element *yaml.RNode) error {
		return f.updateContainer(element)
	})
}

func (f Filter) updateContainer(rn *yaml.RNode) error {
	if err := yaml.ErrorIfInvalid(rn, yaml.MappingNode); err != nil {
		return err
	}

	imageValue, err := rn.GetString("image")
	if err != nil {
		return err
	}

	for _, img := range f.Images {
		if image.IsImageMatched(imageValue, img.Name) {
			return rn.SetMapField(yaml.NewScalarRNode(img.NewPullPolicy), "imagePullPolicy")
		}
	}

	return nil
}

func (f Filter) isOnDenyList(node *yaml.RNode) bool {
	meta, err := node.GetMeta()
	if err != nil {
		// A missing 'meta' field will cause problems elsewhere;
		// ignore it here to keep the signature simple.
		return false
	}
	// Ignore CRDs
	// https://github.com/kubernetes-sigs/kustomize/issues/890
	return meta.Kind == `CustomResourceDefinition`
}
