package kubernetes

import "fmt"

func argocdData(ver string) Application {
	return Application{
		Name:        "argocd",
		Url:         fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", ver),
		Maintainer:  "Dipankar Das",
		Version:     ver,
		InstallType: InstallKubectl,
		KubectlConfig: KubectlOptions{
			metadata: fmt.Sprintf("Argo CD (Ver: %s) is a declarative, GitOps continuous delivery tool for Kubernetes.", ver),
			postInstall: `
			Commands to execute to access Argocd
			$ kubectl get secret -n argocd argocd-initial-admin-secret -o json | jq -r '.data.password' | base64 -d
			$ kubectl port-forward svc/argocd-server -n argocd 8080:443
			and login to http://localhost:8080 with user admin and password from above
			`,
			createNamespace: true,
			namespace:       "argocd",
		},
	}
}

func argocdStandardApp(ver string, withUI bool, clusterAccess bool) ApplicationStack {
	url := fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", ver)
	postInstall := `
	Commands to execute to access Argocd
	$ kubectl get secret -n argocd argocd-initial-admin-secret -o json | jq -r '.data.password' | base64 -d
	$ kubectl port-forward svc/argocd-server -n argocd 8080:443
	and login to http://localhost:8080 with user admin and password from above
`
	if !withUI {
		url = fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/core-install.yaml", ver)
		postInstall = `
https://argo-cd.readthedocs.io/en/stable/operator-manual/core/
`
	}

	// TODO(dipankar): do we still need to create the namespace?
	if !clusterAccess {
		url = fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/namespace-install.yaml", ver)
		postInstall = `
https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#non-high-availability
`
	}

	return ApplicationStack{
		components: []StackComponent{
			{
				kubectl: &KubectlHandler{
					namespace:       "argocd",
					createNamespace: true,
					url:             url,
					version:         ver,
					metadata:        fmt.Sprintf("Argo CD (Ver: %s) is a declarative, GitOps continuous delivery tool for Kubernetes.", ver),
					postInstall:     postInstall,
				},
				handlerType: ComponentTypeKubectl,
			},
		},
		StackType:   StackTypeStandard,
		StackNameID: ArgocdStandardStackID,
		Maintainer:  "github@dipankardas011",
	}
}

// NOTE: always check the compatability
// Refer: https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#tested-versions
// TODO: add it to the ksctl docs

// WARN: the below production section is underdevelopment
// Production has these
// Refer: https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#high-availability
func argocdProductionApp(ver string, clusterAccess bool) ApplicationStack {
	url := fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/ha/install.yaml", ver)
	postInstall := `
	Commands to execute to access Argocd
	$ kubectl get secret -n argocd argocd-initial-admin-secret -o json | jq -r '.data.password' | base64 -d
	$ kubectl port-forward svc/argocd-server -n argocd 8080:443
	and login to http://localhost:8080 with user admin and password from above
`

	// TODO(dipankar): do we still need to create the namespace?
	if !clusterAccess {
		url = fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/ha/namespace-install.yaml", ver)
		postInstall = `
https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#non-high-availability
`
	}

	return ApplicationStack{
		components: []StackComponent{
			{
				kubectl: &KubectlHandler{
					namespace:       "argocd",
					createNamespace: true,
					url:             url,
					version:         ver,
					metadata:        fmt.Sprintf("Argo CD (Ver: %s) is a declarative, GitOps continuous delivery tool for Kubernetes.", ver),
					postInstall:     postInstall,
				},
				handlerType: ComponentTypeKubectl,
			},
		},
		StackType:   StackTypeProduction,
		StackNameID: ArgocdProductionStackID,
		Maintainer:  "github@dipankardas011",
	}
}
