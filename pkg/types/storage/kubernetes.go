package storage

import (
	"fmt"
	"strings"
)

type KubernetesAddons struct {
	Apps []Application `json:"apps" bson:"apps"`
	Cni  Application   `json:"cni" bson:"cni"`
}

type Application struct {
	Name       string               `json:"name" bson:"name"`
	Components map[string]Component `json:"components" bson:"components"`
}

type Component struct {
	Version string `json:"version" bson:"version"`
}

func (c Component) String() string {
	return c.Version
}

func (a Application) String() string {
	if len(a.Name) == 0 {
		return ""
	}
	var components []string
	for _, c := range a.Components {
		x := ""
		if len(c.Version) == 0 {
			continue
		}
		x = c.String()
		components = append(components, x)
	}

	return fmt.Sprintf("%s::[%s]", a.Name, strings.Join(components, ","))
}
