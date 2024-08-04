package stacks

import (
	"github.com/ksctl/ksctl/internal/kubernetes/components"
	meta "github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func ArgocdStandardCICD(params meta.ApplicationParams) meta.ApplicationStack {

	return meta.ApplicationStack{
		Components: map[meta.StackComponentID]meta.StackComponent{
			meta.ArgocdComponentID: components.ArgoCDStandardComponent(
				params.ComponentParams[meta.ArgocdComponentID],
			),
		},

		StkDepsIdx:  []meta.StackComponentID{meta.ArgocdComponentID},
		StackNameID: meta.ArgocdStandardStackID,
		Maintainer:  "github@dipankardas011",
	}
}

// NOTE: always check the compatability
// Refer: https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#tested-versions
// TODO: add it to the ksctl docs

// WARN: the below production section is underdevelopment
// Production has these
// Refer: https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#high-availability
//
// func argocdProductionApp(ver string, clusterAccess bool) ApplicationStack {
// 	url := fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/ha/install.yaml", ver)
// 	postInstall := `
// 	Commands to execute to access Argocd
// 	$ kubectl get secret -n argocd argocd-initial-admin-secret -o json | jq -r '.data.password' | base64 -d
// 	$ kubectl port-forward svc/argocd-server -n argocd 8080:443
// 	and login to http://localhost:8080 with user admin and password from above
// `
//
// 	// TODO(dipankar): do we still need to create the namespace?
// 	if !clusterAccess {
// 		url = fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/ha/namespace-install.yaml", ver)
// 		postInstall = `
// https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#non-high-availability
// `
// 	}
//
// 	return ApplicationStack{
// 		components: []StackComponent{
// 			{
// 				kubectl: &KubectlHandler{
// 					namespace:       "argocd",
// 					createNamespace: true,
// 					url:             url,
// 					version:         ver,
// 					metadata:        fmt.Sprintf("Argo CD (Ver: %s) is a declarative, GitOps continuous delivery tool for Kubernetes.", ver),
// 					postInstall:     postInstall,
// 				},
// 				handlerType: ComponentTypeKubectl,
// 			},
// 		},
// 		StackNameID: ArgocdProductionStackID,
// 		Maintainer:  "github@dipankardas011",
// 	}
// }
