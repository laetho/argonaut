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
	"encoding/base64"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	argonautv1 "github.com/laetho/argonaut/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

	// Reconciliation flow for CloudFlare Resources
	// 1. [ ] Reconcile Argo Tunnel
	// 2. [ ] Reconcile Zone DNS
	// 3. [ ] Reconcile TLS Certificates
	// ?. [ ] Support Load Balancers
	tun, err := r.ReconcileArgoTunnel(ctx, cfc, &argonaut)
	if err != nil {
		fmt.Println("failed to reconcile?")
		return ctrl.Result{Requeue: true, RequeueAfter: 60000000000}, nil
	}
	fmt.Println(tun)

	// Update status on the Argonaut instance.
	if err := r.Status().Update(ctx, &argonaut); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ArgonautReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&argonautv1.Argonaut{}).
		Complete(r)
}

// Get a Cloudflare API instance. Uses login secrets from the secret referenced in the Argonaut spec.
func (r *ArgonautReconciler) CloudflareLogin(ctx context.Context, argonaut *argonautv1.Argonaut) (*cloudflare.API, error) {
	_ = log.FromContext(ctx)
	// Get email and token from secret referenced in argonaut
	var secret v1.Secret
	err := r.Get(ctx, client.ObjectKey{Namespace: argonaut.Spec.CFAuthSecret.Namespace, Name: argonaut.Spec.CFAuthSecret.Name}, &secret)
	if errors.IsNotFound(err) {
		log.FromContext(ctx).Error(err, "Could not find Secret with credentials for Cloudflare API Login: ")
		return nil, err
	}

	token := secret.Data["token"]
	accountid := secret.Data["accountid"]
	if len(token) == 0 {
		return nil, fmt.Errorf("token is missing or has zero length")
	}

	cfc, err := cloudflare.NewWithAPIToken(string(token))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	cfc.AccountID = string(accountid)
	return cfc, nil
}

// Ensures that the Cloudflare Argo Tunnel exists. Will be created if does not exist.
// The Argo Tunnel name will be the name of the Argonaut resource.
func (r *ArgonautReconciler) ReconcileArgoTunnel(ctx context.Context, cfc *cloudflare.API, argonaut *argonautv1.Argonaut) (*cloudflare.ArgoTunnel, error) {

	tun, err := r.GetArgoTunnel(ctx, cfc, argonaut)
	if err != nil {
		return nil, err
	}

	// Write TunnelId to Status
	if tun.ID != argonaut.Status.TunnelId {
		argonaut.Status.TunnelId = tun.ID
	}

	if len(tun.ID) == 0 {
		// Create Tunnel
		tun, err = r.CreateArgoTunnel(ctx, cfc, argonaut)
		if err != nil {
			return nil, err
		}
		fmt.Println(tun)
	}

	return &tun, nil
}

// func (r *ArgonautReconciler) ReconcileArgoTunnelSecret(ctx context.Context, cfc, argonaut *argonautv1.Argonaut)

// Fetch a Argo Tunnel from the Cloudflare API.
func (r *ArgonautReconciler) GetArgoTunnel(ctx context.Context, cfc *cloudflare.API, argonaut *argonautv1.Argonaut) (cloudflare.ArgoTunnel, error) {

	tuns, err := cfc.ArgoTunnels(ctx, cfc.AccountID)
	if err != nil {
		return cloudflare.ArgoTunnel{}, err
	}

	for _, tun := range tuns {
		if tun.Name == argonaut.Spec.ArgoTunnelName {
			return tun, nil
		}
	}

	return cloudflare.ArgoTunnel{}, nil
}

// Create a Argo Tunnel using the Cloudflare API
func (r *ArgonautReconciler) CreateArgoTunnel(ctx context.Context, cfc *cloudflare.API, argonaut *argonautv1.Argonaut) (cloudflare.ArgoTunnel, error) {

	tun, err := cfc.CreateArgoTunnel(ctx, cfc.AccountID, argonaut.Spec.ArgoTunnelName, base64.StdEncoding.EncodeToString([]byte("SuperSecretStringGeneratorHere")))
	if err != nil {
		return cloudflare.ArgoTunnel{}, err
	}
	return tun, nil
}

// Create or Update a Secret with TunnelID and TunnelSecret.
func (r *ArgonautReconciler) CreateOrUpdateArgoTunnelSecret(ctx context.Context, argonaut *argonautv1.Argonaut, tun *cloudflare.ArgoTunnel) (*v1.Secret, error) {
	var argoTunnelSecret v1.Secret

	argoTunnelSecret.Name = argonaut.Name
	argoTunnelSecret.Namespace = argonaut.Namespace

	if err := r.Create(ctx, &argoTunnelSecret); err != nil {
		return nil, nil
	}
	return &argoTunnelSecret, nil
}

// Deletes an Argo Tunnel using the Cloudflare API
func (r *ArgonautReconciler) DeleteArgoTunnel(ctx context.Context, cfc *cloudflare.API, tun *cloudflare.ArgoTunnel) error {
	if err := cfc.DeleteArgoTunnel(ctx, cfc.AccountID, tun.ID); err != nil {
		return err
	}
	log.FromContext(ctx).Info("deleted Argo Tunnel", tun.ID, tun.Name)
	return nil
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
