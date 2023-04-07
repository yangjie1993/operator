package controllers

import (
	"github.com/yangjie1993/operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func MutateDeployment(app *v1beta1.MyApp, deploy *appsv1.Deployment) {
	labels := map[string]string{"app": app.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	deploy.Spec = appsv1.DeploymentSpec{
		Replicas: app.Spec.Size,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: newContainers(app),
			},
		},
		Selector: selector,
	}
}

func NewDeploy(app *v1beta1.MyApp) *appsv1.Deployment {
	labels := map[string]string{"app": app.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			//
			OwnerReferences: makeOwnerRefer(app),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: app.Spec.Size,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: newContainers(app),
				},
			},
			Selector: selector,
		},
	}
}

func makeOwnerRefer(app *v1beta1.MyApp) []metav1.OwnerReference {
	//OwnerReferences
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(app, schema.GroupVersionKind{
			Kind:    v1beta1.Kind,
			Group:   v1beta1.GroupVersion.Group,
			Version: v1beta1.GroupVersion.Version,
		}),
	}
}

func newContainers(app *v1beta1.MyApp) []corev1.Container {

	containerPorts := []corev1.ContainerPort{}
	for _, svcPort := range app.Spec.Ports {
		cport := corev1.ContainerPort{}
		cport.ContainerPort = svcPort.TargetPort.IntVal

		containerPorts = append(containerPorts, cport)
	}
	return []corev1.Container{
		{
			Name:            app.Name,
			Image:           app.Spec.Image,
			Resources:       app.Spec.Resources,
			Ports:           containerPorts,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env:             app.Spec.Envs,
		},
	}

}

func MutateService(app *v1beta1.MyApp, svc *corev1.Service) {
	svc.Spec = corev1.ServiceSpec{
		ClusterIP: svc.Spec.ClusterIP,
		Ports:     app.Spec.Ports,
		Type:      corev1.ServiceTypeNodePort,
		Selector: map[string]string{
			"myapp": app.Name,
		},
	}
}

func NewService(app *v1beta1.MyApp) *corev1.Service {

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            app.Name,
			Namespace:       app.Namespace,
			OwnerReferences: makeOwnerRefer(app),
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: app.Spec.Ports,
			Selector: map[string]string{
				"app": app.Name,
			},
		},
	}
}
