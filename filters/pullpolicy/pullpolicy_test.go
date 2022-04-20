package pullpolicy

import (
	"strings"
	"testing"

	filtertest "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestFilter_Filter(t *testing.T) {
	testCases := map[string]struct {
		filter         Filter
		input          string
		expectedOutput string
		expectErr      bool
	}{
		"add missing policy in pod": {
			filter: Filter{
				Images: []imageEntry{{
					Name:          "nginx",
					NewPullPolicy: "Always",
				}},
			},
			input: `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.14.2
`,
			expectedOutput: `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.14.2
    imagePullPolicy: Always
`,
		},
		"missing image": {
			filter: Filter{
				Images: []imageEntry{{
					Name:          "nginx",
					NewPullPolicy: "Always",
				}},
			},
			input: `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
`,
			expectedOutput: "",
			expectErr:      true,
		},
		"containers node not a sequence": {
			filter: Filter{
				Images: []imageEntry{{
					Name:          "nginx",
					NewPullPolicy: "Always",
				}},
			},
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: whatever
spec:
  containers: 'not what you expected'
`,
			expectedOutput: "",
			expectErr:      true,
		},
		"container node wrong type": {
			filter: Filter{
				Images: []imageEntry{{
					Name:          "nginx",
					NewPullPolicy: "Always",
				}},
			},
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: whatever
spec:
  containers:
  - 'not what you expected'
`,
			expectedOutput: "",
			expectErr:      true,
		},
		"no matches on image in pod": {
			filter: Filter{
				Images: []imageEntry{{
					Name:          "nginx",
					NewPullPolicy: "Always",
				}},
			},
			input: `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: notnginx:1.14.2
`,
			expectedOutput: `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: notnginx:1.14.2
`,
		},
		"update multiple containers and initContainers in pod": {
			filter: Filter{
				Images: []imageEntry{{
					Name:          "nginx",
					NewPullPolicy: "Always",
				}},
			},
			input: `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  initContainers:
  - name: nginx-init
    image: nginx:1.14.2
    imagePullPolicy: IfNotPresent
  containers:
  - name: nginx-specific
    image: nginx:1.14.2
    imagePullPolicy: IfNotPresent
  - name: nginx-latest
    image: nginx:latest
    imagePullPolicy: Always
`,
			expectedOutput: `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  initContainers:
  - name: nginx-init
    image: nginx:1.14.2
    imagePullPolicy: Always
  containers:
  - name: nginx-specific
    image: nginx:1.14.2
    imagePullPolicy: Always
  - name: nginx-latest
    image: nginx:latest
    imagePullPolicy: Always
`,
		},
		"add missing policy to deployment": {
			filter: Filter{
				Images: []imageEntry{{
					Name:          "nginx",
					NewPullPolicy: "Always",
				}},
			},
			input: `
group: apps
apiVersion: v1
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
`,
			expectedOutput: `
group: apps
apiVersion: v1
metadata:
  name: nginx
spec:
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        imagePullPolicy: Always
`,
		},
		"custom field specs": {
			filter: Filter{
				Images: []imageEntry{{
					Name:          "nginx",
					NewPullPolicy: "Always",
				}},
				FsSlice: []types.FieldSpec{
					{
						Path:               "spec/foos[]",
						CreateIfNotPresent: true,
					},
				},
			},
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: nginx
spec:
  foos:
  - name: nginx
    image: nginx:1.14.2
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: nginx
spec:
  foos:
  - name: nginx
    image: nginx:1.14.2
    imagePullPolicy: Always
`,
		},
		// Based on test in Kustomize repository.
		// https://github.com/kubernetes-sigs/kustomize/blob/9d5491c2e20c23c7a9b3d5a055d2aa24cc14bede/api/filters/imagetag/imagetag_test.go#L26
		"ignore CustomResourceDefinition": {
			input: `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: whatever
spec:
  containers:
  - image: whatever
`,
			expectedOutput: `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: whatever
spec:
  containers:
  - image: whatever
`,
			filter: Filter{
				Images: []imageEntry{{
					Name:          "whatever",
					NewPullPolicy: "Always",
				}},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			filter := tc.filter

			// Run the Defaulter.
			err := filter.Default()
			if err != nil {
				t.Errorf("%q: Defaulter error: %v", tn, err)
			}

			expectedOutput := strings.TrimSpace(tc.expectedOutput)
			actualOutput, actualErr := filtertest.RunFilterE(t, tc.input, filter)
			actualOutput = strings.TrimSpace(actualOutput)
			wasErr := actualErr != nil

			if tc.expectErr != wasErr {
				t.Errorf("%q: expectErr=%v but wasErr=%v", tn, tc.expectErr, wasErr)
			}
			if expectedOutput != actualOutput {
				t.Errorf("%q: actual output doesn't match expected\n---\n# Expected\n%s\n---\n# Actual\n%s", tn, expectedOutput, actualOutput)
			}
		})
	}
}
