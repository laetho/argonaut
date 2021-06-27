package controllers

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
	argonautv1 "github.com/laetho/argonaut/api/v1beta1"
)

// Reconcile hostnames found in Argonaut instance with CloudFlare DNS
func (r *ArgonautReconciler) ReconcileDNS(ctx context.Context, cfc *cloudflare.API, argonaut *argonautv1.Argonaut, tun cloudflare.ArgoTunnel) error {
	// Find all hostnames in ingress, reconcile
	return nil
}
