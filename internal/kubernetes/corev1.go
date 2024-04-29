package kubernetes

import (
	"context"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (this *Kubernetes) configMapApply(o *corev1.ConfigMap, ns string) error {

	_, err := this.clientset.
		CoreV1().
		ConfigMaps(ns).
		Create(context.Background(), o, metav1.CreateOptions{})

	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = this.clientset.
				CoreV1().
				ConfigMaps(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}

		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) configMapDelete(o *corev1.ConfigMap, ns string) error {

	err := this.clientset.
		CoreV1().
		ConfigMaps(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func (this *Kubernetes) serviceDelete(o *corev1.Service, ns string) error {

	err := this.clientset.
		CoreV1().
		Services(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func (this *Kubernetes) serviceApply(o *corev1.Service, ns string) error {

	_, err := this.clientset.
		CoreV1().
		Services(ns).
		Create(context.Background(), o, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = this.clientset.
				CoreV1().
				Services(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}

		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) namespaceCreate(ns string) error {

	if _, err := this.clientset.
		CoreV1().
		Namespaces().
		Create(context.Background(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.
				CoreV1().
				Namespaces().
				Update(context.Background(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}, metav1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) namespaceDelete(ns string) error {

	if err := this.clientset.
		CoreV1().
		Namespaces().
		Delete(context.Background(), ns, metav1.DeleteOptions{
			GracePeriodSeconds: func() *int64 {
				v := int64(0)
				return &v
			}(),
		}); err != nil {
		return log.NewError(err.Error())
	}

	for i := 0; i < int(consts.CounterMaxRetryCount); {
		_, err := this.clientset.
			CoreV1().
			Namespaces().
			Get(context.Background(), ns, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) || apierrors.IsGone(err) {
				log.Debug("Namespace deleted", "namespace", ns)
				break
			} else {
				return log.NewError("Failed to get namespace", "namespace", ns, "err", err)
			}
		}

		// Namespace still exists, wait and check again
		log.Warn("Namespace still deleting", "namespace", ns, "retry", i, "threshold", int(consts.CounterMaxRetryCount))
		time.Sleep(5 * time.Second)
		i++
	}
	return nil
}

func (this *Kubernetes) secretDelete(o *corev1.Secret, ns string) error {

	err := this.clientset.
		CoreV1().
		Secrets(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func (this *Kubernetes) secretApply(o *corev1.Secret, ns string) error {

	_, err := this.clientset.
		CoreV1().
		Secrets(ns).
		Create(context.Background(), o, metav1.CreateOptions{})

	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = this.clientset.
				CoreV1().
				Secrets(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}

		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) PodApply(o *corev1.Pod, ns string) error {

	_, err := this.clientset.
		CoreV1().
		Pods(ns).
		Create(context.Background(), o, metav1.CreateOptions{})

	if err != nil {
		if apierrors.IsAlreadyExists(err) {

			_, err = this.clientset.
				CoreV1().
				Pods(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}

		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) PodDelete(o *corev1.Pod, ns string) error {

	err := this.clientset.
		CoreV1().
		Pods(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})

	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func (this *Kubernetes) serviceAccountDelete(o *corev1.ServiceAccount, ns string) error {

	err := this.clientset.
		CoreV1().
		ServiceAccounts(ns).
		Delete(context.Background(), o.Name, metav1.DeleteOptions{})
	if err != nil {
		return log.NewError(err.Error())
	}
	return nil
}

func (this *Kubernetes) serviceAccountApply(o *corev1.ServiceAccount, ns string) error {

	_, err := this.clientset.
		CoreV1().
		ServiceAccounts(ns).
		Create(context.Background(), o, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			_, err = this.clientset.
				CoreV1().
				ServiceAccounts(ns).
				Update(context.Background(), o, metav1.UpdateOptions{})
			if err != nil {
				return log.NewError(err.Error())
			}
		} else {
			return log.NewError(err.Error())
		}
	}
	return nil
}

func (this *Kubernetes) nodesList() (*corev1.NodeList, error) {
	return this.clientset.
		CoreV1().
		Nodes().
		List(context.Background(), metav1.ListOptions{})
}

func (this *Kubernetes) nodeDelete(nodeName string) error {
	return this.clientset.
		CoreV1().
		Nodes().
		Delete(context.Background(), nodeName, metav1.DeleteOptions{})
}
