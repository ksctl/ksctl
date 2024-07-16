package stacks

import "github.com/ksctl/ksctl/internal/kubernetes/metadata"

func ConvertDepsFromMapToList(in metadata.ApplicationStack) (components []metadata.StackComponent) {
	for k, v := range in.Components {
	}
	return
}
