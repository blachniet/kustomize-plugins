package main

import (
	"os"

	"github.com/blachniet/kustomize-plugins/filters/pullpolicy"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func main() {
	p := &framework.VersionedAPIProcessor{FilterProvider: framework.GVKFilterMap{
		"PullPolicyTransformer": {
			"k8s.blachniet.com/v1alpha1": &pullpolicy.Filter{},
		},
	}}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
