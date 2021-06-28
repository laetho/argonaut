package controllers

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
	argonautv1 "github.com/laetho/argonaut/api/v1beta1"
)

// Reconcile hostnames found in Argonaut instance with CloudFlare DNS
func (r *ArgonautReconciler) ReconcileDNS(ctx context.Context, cfc *cloudflare.API, argonaut *argonautv1.Argonaut, tun *cloudflare.ArgoTunnel) error {
	// Find all hostnames in ingress, reconcile records.
	zone, err := r.ReconcileZone(ctx, cfc, argonaut)
	if err != nil {
		return err
	}

	records, err := r.GetDNSRecords(ctx, cfc, zone)
	if err != nil {
		return err
	}

	for _, ingress := range argonaut.Spec.Ingress {
		if inDNSRecords(records, ingress.Hostname) {
			// update
		} else {
			// create
		}
	}
	return nil
}

// Fetch all DNS records for a Zone. We're only interestedin CNAME's for tunnels.
func (r *ArgonautReconciler) GetDNSRecords(ctx context.Context, cfc *cloudflare.API, zoneid string) ([]cloudflare.DNSRecord, error) {
	records, err := cfc.DNSRecords(ctx, zoneid, cloudflare.DNSRecord{Type: "CNAME"})
	if err != nil {
		return records, err
	}
	return records, nil
}

func inDNSRecords(records []cloudflare.DNSRecord, item string) bool {

	return true
}
