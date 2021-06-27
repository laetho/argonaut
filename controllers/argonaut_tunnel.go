package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	argonautv1 "github.com/laetho/argonaut/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

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
func (r *ArgonautReconciler) ReconcileArgoTunnelSecret(ctx context.Context, argonaut *argonautv1.Argonaut, tun *cloudflare.ArgoTunnel) (*v1.Secret, error) {
	var argoTunnelSecret v1.Secret

	argoTunnelSecret.Name = argonaut.Name
	argoTunnelSecret.Namespace = argonaut.Namespace
	argoTunnelSecret.StringData["secret"] = tun.Secret
	argoTunnelSecret.StringData["tunnelid"] = tun.ID

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
