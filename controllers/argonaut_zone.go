package controllers

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	argonautv1 "github.com/laetho/argonaut/api/v1beta1"
	"strings"
	//	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconcile Zones
func (r *ArgonautReconciler) ReconcileZone(ctx context.Context, cfc *cloudflare.API, argonaut *argonautv1.Argonaut) error {
	if !r.ZoneExists(ctx, cfc, HostnameToZone(argonaut.Spec.Ingress[0].Hostname)) {
		return fmt.Errorf("zone does not exist")
	}
	return nil
}

// Check if a DNS Zone exists.
func (r *ArgonautReconciler) ZoneExists(ctx context.Context, cfc *cloudflare.API, name string) bool {

	zones, err := cfc.ListZones(ctx, name)
	if err != nil {
		return false
	}

	fmt.Println(zones)

	return true

}

// Helper function to turn a hostname into a Zone. Example blah.example.com into example.com.
func HostnameToZone(hostname string) (zone string) {
	parts := strings.Split(hostname, ".")
	nparts := len(parts)
	zone = parts[nparts-2] + "." + parts[nparts-1]
	return
}
