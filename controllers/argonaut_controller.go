/*
Copyright 2021 The Argonaut authors.

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
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	argonautv1 "github.com/laetho/argonaut/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	// Stable BaseURL for Cloudflare API endpoints.
	CloudflareBaseURL = "https://api.cloudflare.com/client/v4"
)

// ArgonautReconciler reconciles a Argonaut object
type ArgonautReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=argonaut.metalabs.no,resources=argonauts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argonaut.metalabs.no,resources=argonauts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=argonaut.metalabs.no,resources=argonauts/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ArgonautReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var argonaut argonautv1.Argonaut
	if err := r.Get(ctx, req.NamespacedName, &argonaut); err != nil {
		log.FromContext(ctx).Error(err, "unable to fetch Argonaut resource")
		// Potentially handle removal?
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	cfc, err := r.CloudflareLogin(ctx, &argonaut)
	if err != nil {
		return ctrl.Result{}, err
	}

	fmt.Println(cfc.APIEmail, cfc.AccountID)

	// Find endpoints to create config for.
	//eps := r.EndpointsLists(ctx, &argonaut)

	tuns, err := cfc.ArgoTunnels(ctx, cfc.AccountID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(tuns)

	/*
		dep := v1beta1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       v1beta1.DeploymentSpec{},
			Status:     v1beta1.DeploymentStatus{},
		}
	*/

	// Find zone
	// DNS
	// Tunnel
	// Create or Update Pod

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ArgonautReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&argonautv1.Argonaut{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				// Supress update events on same generation of Argonaut resource.
				oldGen := e.ObjectOld.GetGeneration()
				newGen := e.ObjectNew.GetGeneration()
				fmt.Println(oldGen, newGen)
				if newGen == oldGen {
					return false
				}
				return true
			},
		}).
		Complete(r)
}

func (r *ArgonautReconciler) CloudflareLogin(ctx context.Context, argonaut *argonautv1.Argonaut) (*cloudflare.API, error) {
	// Get email and token from secret referenced in argonaut
	var secret v1.Secret
	err := r.Get(ctx, client.ObjectKey{Namespace: argonaut.Spec.CFAuthSecret.Namespace, Name: argonaut.Spec.CFAuthSecret.Name}, &secret)
	if errors.IsNotFound(err) {
		fmt.Println("Could not find Secret with credentials for Cloudflare API Login: ", err)
		return nil, err
	}

	token := secret.Data["token"]
	email := secret.Data["email"]
	if len(token) == 0 {
		return nil, fmt.Errorf("Token is missing or has zero length")
	}

	cfc, err := cloudflare.New(string(token), string(email), cloudflare.BaseURL(CloudflareBaseURL))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return cfc, nil
}

// Get a a slice of EndpointsList for an Argonaut resource
func (r *ArgonautReconciler) EndpointsLists(ctx context.Context, argonaut *argonautv1.Argonaut) []v1.EndpointsList {
	var eps []v1.EndpointsList
	for _, h := range argonaut.Spec.Ingress {
		var el v1.EndpointsList
		err := r.List(ctx, &el, client.MatchingLabels(h.EndpointsSelector.MatchLabels))
		if errors.IsNotFound(err) {
			fmt.Println("Did not find EndpointsList with selector:", h.EndpointsSelector.MatchLabels)
		}
		eps = append(eps, el)
	}
	return eps
}
