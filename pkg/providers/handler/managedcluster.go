package handler

func (kc *Controller) CreateManagedCluster() (bool, bool, error) {

	if kc.b.IsLocalProvider(kc.p) {
		if err := kc.p.Cloud.Name(kc.p.Metadata.ClusterName + "-ksctl-managed-net").NewNetwork(); err != nil {
			return false, false, err
		}
	}

	managedClient := kc.p.Cloud.Name(kc.p.Metadata.ClusterName + "-ksctl-managed")

	managedClient = managedClient.VMType(kc.p.Metadata.ManagedNodeType)

	externalApps := managedClient.Application(
		func() (apps []string) {
			for _, ss := range kc.p.Metadata.Applications {
				apps = append(apps, ss.StackName)
			}
			return apps
		}())

	externalCNI := managedClient.CNI(kc.p.Metadata.CNIPlugin.StackName)

	managedClient = managedClient.ManagedK8sVersion(kc.p.Metadata.K8sVersion)

	if managedClient == nil {
		return externalApps, externalCNI, kc.l.NewError(kc.ctx, "invalid k8s version")
	}

	if err := managedClient.NewManagedCluster(kc.p.Metadata.NoMP); err != nil {
		return externalApps, externalCNI, err
	}
	return externalApps, externalCNI, nil
}

func (kc *Controller) DeleteManagedCluster() error {

	if err := kc.p.Cloud.DelManagedCluster(); err != nil {
		return err
	}

	if kc.b.IsLocalProvider(kc.p) {
		if err := kc.p.Cloud.DelNetwork(); err != nil {
			return err
		}
	}
	return nil
}
