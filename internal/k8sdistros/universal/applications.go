package universal

import (
	"fmt"
	"io"

	"net/http"
	"strings"

	"github.com/kubesimplify/ksctl/pkg/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

type Application struct {
	Url         string
	Namespace   string
	Version     string
	Metadata    string
	Maintainer  string
	PostInstall string
}

var (
	apps map[string]func() Application
)

func initApps() {
	apps = map[string]func() Application{
		// "cilium":  {},
		// "flannel": {},
		"argocd": argocdData,
	}
}

func argocdData() Application {
	return Application{
		Url:        "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml",
		Namespace:  "argocd",
		Maintainer: "Dipankar Das",
		Version:    "Latest stable version",
		Metadata:   "Argo CD is a declarative, GitOps continuous delivery tool for Kubernetes.",
		PostInstall: `
Commands to execute to access Argocd
	kubectl get secret -n argocd argocd-initial-admin-secret -o json | jq -r '.data.password' | base64 -d
	kubectl port-forward svc/argocd-server -n argocd 8080:443
and login to http://localhost:8080 with user admin and password from above
		`,
	}
}

func GetApps(storage resources.StorageFactory, name string) (Application, error) {
	if apps == nil {
		return Application{}, fmt.Errorf("[kubernetes] app variable not initalized")
	}

	val, present := apps[name]

	if !present {
		return Application{}, fmt.Errorf("[kubernetes] app not found %s", name)
	}
	return val(), nil
}

func getManifests(app Application) ([]string, error) {

	// Get the manifest
	resp, err := http.Get(app.Url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Split the manifest into individual resources
	resources := strings.Split(string(body), "---")
	if err := apiextensionsv1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}
	return resources, nil
}

func (this *Kubernetes) InstallApplications(apps []string) error {

	for _, app := range apps {
		if err := installApplication(this, app); err != nil {
			return err
		}
	}

	return nil
}

func installApplication(client *Kubernetes, app string) error {

	appStruct, err := GetApps(client.StorageDriver, app)
	if err != nil {
		return err
	}

	resources, err := getManifests(appStruct)
	if err != nil {
		return err
	}

	if err := client.namespaceCreate(appStruct.Namespace); err != nil {
		return err
	}

	for _, resource := range resources {
		// fmt.Println(resource)
		// Decode the resource into an unstructured object
		decUnstructured := scheme.Codecs.UniversalDeserializer().Decode

		obj, _, err := decUnstructured([]byte(resource), nil, nil)
		if err != nil {
			return err
		}

		var errRes error

		switch o := obj.(type) {

		case *apiextensionsv1.CustomResourceDefinition:
			errRes = client.apiextensionsApply(o)

		case *appsv1.Deployment:

			fmt.Printf("Deployment %T {%s}\n", o, o.Name)
			errRes = client.deploymentApply(o, appStruct.Namespace)

		case *corev1.Service:
			fmt.Printf("service %T {%s}\n", o, o.Name)
			errRes = client.serviceApply(o, appStruct.Namespace)

		case *corev1.ServiceAccount:
			fmt.Printf("serviceaccount %T {%s}\n", o, o.Name)
			errRes = client.serviceaccountApply(o, appStruct.Namespace)

		case *corev1.ConfigMap:
			fmt.Printf("configmap %T {%s}\n", o, o.Name)
			errRes = client.configMapApply(o, appStruct.Namespace)

		case *corev1.Secret:
			fmt.Printf("Secret %T {%s}\n", o, o.Name)
			errRes = client.secretApply(o, appStruct.Namespace)

		case *appsv1.StatefulSet:
			fmt.Printf("Statefulset %T {%s}\n", o, o.Name)
			errRes = client.statefulsetApply(o, appStruct.Namespace)

		case *rbacv1.ClusterRole:
			fmt.Printf("ClusterRole %T {%s}\n", o, o.Name)
			errRes = client.clusterroleApply(o)

		case *rbacv1.ClusterRoleBinding:
			fmt.Printf("ClusterRoleBinding %T {%s}\n", o, o.Name)
			errRes = client.clusterrolebindingApply(o)

		case *rbacv1.Role:
			fmt.Printf("Role %T {%s}\n", o, o.Name)
			errRes = client.roleApply(o, appStruct.Namespace)

		case *rbacv1.RoleBinding:
			fmt.Printf("RoleBinding %T {%s}\n", o, o.Name)
			errRes = client.rolebindingApply(o, appStruct.Namespace)

		case *networkingv1.NetworkPolicy:
			fmt.Printf("NetworkPolicy %T {%s}\n", o, o.Name)
			errRes = client.netpolicyApply(o, appStruct.Namespace)

		default:
			fmt.Printf("unexpected type %T\n", o)
		}

		if errRes != nil {
			return errRes
		}
	}

	client.StorageDriver.Logger().Success("[kubernetes] Installed", app, "\n", appStruct.PostInstall)
	return nil
}
