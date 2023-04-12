/*
Copyright 2023 yj.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	appv1beta1 "github.com/yangjie1993/operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	oldSpecAnnotation = "old/spec"
)

// MyAppReconciler reconciles a MyApp object
type MyAppReconciler struct {
	Log logr.Logger
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=service,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.yj.io,resources=myapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.yj.io,resources=myapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *MyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = log.FromContext(ctx)
	log := log.FromContext(ctx)
	//log := r.Log.WithValues("myapp", req.NamespacedName)
	//ctx := context.Background()

	//首先我们获取MyApp实例
	var myapp appv1beta1.MyApp
	err := r.Client.Get(ctx, req.NamespacedName, &myapp)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		// 再删除一个不存在的对象的时候，可能会报not-found的错误
		// 这种情况不需要重新入队列排修复
		return ctrl.Result{}, nil
	}

	//当前的对象标记为了删除
	if myapp.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}
	//log.Info("fetch myapp objects", "myapp", myapp)
	//log.Info("fetch myapp objects", "myapp22", 22)
	//调谐，获取到当前的一个状态，然后和我们期望的状态进行对比是不是就可以
	//CreateUpdate Deployment
	var deploy appsv1.Deployment
	deploy.Name = myapp.Name
	deploy.Namespace = myapp.Namespace
	or, err := ctrl.CreateOrUpdate(ctx, r.Client, &deploy, func() error {
		//调谐必须在这个函数中去实现
		MutateDeployment(&myapp, &deploy)
		return controllerutil.SetControllerReference(&myapp, &deploy, r.Scheme)
	})

	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("CreateOrUpdate", "Deployment", or)
	// CreateOrUpdate service
	var svc corev1.Service
	svc.Name = myapp.Name
	svc.Namespace = myapp.Namespace
	or, err = ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		MutateService(&myapp, &svc)
		return controllerutil.SetControllerReference(&myapp, &svc, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("CreateOrUpdate", "Service", or)

	//// 如果不存在关联的资源，是不是应该去创建
	//// 如果存在关联的资源，是不是也要判断是否需要更新
	//deploy := &appsv1.Deployment{}
	//if err := r.Client.Get(ctx, req.NamespacedName, deploy); err != nil && errors.IsNotFound(err) {
	//	//关联Annotation
	//	data, err := json.Marshal(myapp.Spec)
	//	if err != nil {
	//		return ctrl.Result{}, err
	//	}
	//	if myapp.Annotations != nil {
	//		myapp.Annotations[oldSpecAnnotation] = string(data)
	//	} else {
	//		myapp.Annotations = map[string]string{
	//			oldSpecAnnotation: string(data),
	//		}
	//	}
	//	//重新更新myapp
	//	retry.RetryOnConflict(retry.DefaultRetry, func() error {
	//		return r.Client.Update(ctx, &myapp)
	//	})
	//	if err != nil {
	//		return ctrl.Result{}, err
	//	}
	//
	//	//Deployment 不存在，创建关联的资源
	//	newDeploy := NewDeploy(&myapp)
	//	if err := r.Client.Create(ctx, newDeploy); err != nil {
	//		return ctrl.Result{}, err
	//	}
	//	//直接创建service
	//	newService := NewService(&myapp)
	//	if err := r.Client.Create(ctx, newService); err != nil {
	//		return ctrl.Result{}, err
	//	}
	//	// 创建成功
	//	return ctrl.Result{}, nil
	//}
	//
	//// yaml -> old yaml 我们可以从annotations里面去获取
	//// 当前规范与旧的对象不一致，则需要更新
	//
	//oldspec := appv1beta1.MyAppSpec{}
	//if err := json.Unmarshal([]byte(myapp.Annotations[oldSpecAnnotation]), &oldspec); err != nil {
	//	return ctrl.Result{}, err
	//}
	//newDeploy := NewDeploy(&myapp)
	//oldDeploy := appsv1.Deployment{}
	//if err := r.Client.Get(ctx, req.NamespacedName, &oldDeploy); err != nil {
	//	return ctrl.Result{}, err
	//}
	//oldDeploy.Spec = newDeploy.Spec
	////正常就应该直接去更新oldDeploy
	////注意，一般情况不会直接调用Update进行更新
	//if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
	//	if err := r.Client.Update(ctx, &oldDeploy); err != nil {
	//		return err
	//	}
	//	return nil
	//}); err != nil {
	//	return ctrl.Result{}, nil
	//}
	//// 更新Service
	//newService := NewService(&myapp)
	//oldService := corev1.Service{}
	//if err := r.Client.Get(ctx, req.NamespacedName, &oldService); err != nil {
	//	return ctrl.Result{}, err
	//}
	////需要指定ClusterIP为之前的，否则更新会报错
	//newService.Spec.ClusterIP = oldService.Spec.ClusterIP
	//oldService.Spec = newService.Spec
	//if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
	//	if err := r.Client.Update(ctx, &oldService); err != nil {
	//		return err
	//	}
	//	return nil
	//}); err != nil {
	//	return ctrl.Result{}, err
	//}
	//return ctrl.Result{}, nil
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.MyApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
