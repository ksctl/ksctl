package cli

type Builder struct {
	Client interface{}
}
type CobraCmd struct {
	ClusterName string
	Region      string
	Client      Builder
	Version     string
}
