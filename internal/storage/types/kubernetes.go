package types

type KubernetesAddons struct {
	Apps []Application `json:"apps" bson:"apps"`
	Cni  Application   `json:"cni" bson:"cni"`
}

type Application struct {
	Name    string `json:"name" bson:"name"`
	Version string `json:"version" bson:"version"`
}
