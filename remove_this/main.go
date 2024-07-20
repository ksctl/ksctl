package main

import (
	"fmt"
	"github.com/gookit/goutil/dump"
	"gopkg.in/yaml.v3"
	"os"
)

type StackSpec struct {
	Name      string         `yaml:"stackName"`
	Overrides map[string]any `yaml:"overrides"`
}

type Stack struct {
	Spec StackSpec `yaml:"spec"`
}

func main() {
	b, err := os.ReadFile("deploy.yaml")
	if err != nil {
		panic(err)
	}

	data := new(Stack)
	if _err := yaml.Unmarshal(b, &data); _err != nil {
		panic(_err)
	}

	fmt.Println("StackName:", data.Spec.Name)
	for componentID, fields := range data.Spec.Overrides {
		fmt.Println("ComponentID:", componentID)

		switch o := fields.(type) {
		case map[string]any:
			fmt.Println("OverriddenFields map[string]any")
			if v, ok := o["version"].(string); ok {
				fmt.Println("version:", v)
			}
			if v, ok := o["noUI"].(bool); ok {
				fmt.Println("noUI:", v)
			}
			dump.Println(o)
		case string:
			fmt.Println("OverriddenFields[string]")
			fmt.Println(o)
		case bool:
			fmt.Println("OverriddenFields[bool]")
			fmt.Println(o)
		}
	}
}
