package storage

import "fmt"

type KubernetesAddons struct {
	Apps []Application `json:"apps" bson:"apps"`
	Cni  Application   `json:"cni" bson:"cni"`
}

type Application struct {
	Name    string `json:"name" bson:"name"`
	Version string `json:"version" bson:"version"`
}

func (a Application) String() string {
	if len(a.Name) == 0 {
		return ""
	}
	return fmt.Sprintf("%s@%s", a.Name, a.Version)
}
