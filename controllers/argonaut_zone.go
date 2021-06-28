package controllers

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
	argonautv1 "github.com/laetho/argonaut/api/v1beta1"
	"strings"
	//	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconcile Zones
func (r *ArgonautReconciler) ReconcileZone(ctx context.Context, cfc *cloudflare.API, argonaut *argonautv1.Argonaut) (string, error) {
	zoneid, err := r.ZoneExists(ctx, cfc, HostnameToZone(argonaut.Spec.Ingress[0].Hostname))
	if err != nil {
		return "", err
	}
	return zoneid, nil
}

// Check if a DNS Zone exists.
func (r *ArgonautReconciler) ZoneExists(ctx context.Context, cfc *cloudflare.API, name string) (string, error) {
	zoneid, err := cfc.ZoneIDByName(name)
	if err != nil {
		return "", err
	}
	return zoneid, nil
}

// Helper function to turn a hostname into a Zone. Example blah.example.com into example.com.
func HostnameToZone(hostname string) (zone string) {
	parts := strings.Split(hostname, ".")
	nparts := len(parts)
	zone = parts[nparts-2] + "." + parts[nparts-1]
	return
}
