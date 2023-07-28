package kubernetes

// func NewController(b *k8s.ClientBuilder, state cloud.CloudResourceState) {
// 	// TODO: Which one to call controller will decide
// 	var abcd k8s.ControllerInterface

// 	fmt.Println("[CONTROLLER]: Recieved Cloud Resource State", state)

// 	switch b.K8sDistro {
// 	case "k3s":
// 		abcd = k3s.WrapK8sControllerBuilder(b)
// 	case "kubeadm":
// 		abcd = kubeadm.WrapK8sControllerBuilder(b)
// 	}
// 	fmt.Printf("\tReq for HA: %v\n\n", b.IsHA)
// 	_, _ = abcd.GetKubeconfig()
// 	abcd.HydrateCloudState(state)
// 	// abcd.SetupLoadBalancer()
// 	// _, _ = abcd.SetupDatastore()
// }
// func WrapK8sEngineBuilder(b *resources.Builder) *k8s.ClientBuilder {
// 	api := (*k8s.ClientBuilder)(b)
// 	return api
// }
