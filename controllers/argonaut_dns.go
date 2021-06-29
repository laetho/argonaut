package controllers

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	argonautv1 "github.com/laetho/argonaut/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
		exists, record := inDNSRecords(records, ingress.Hostname)
		if exists {
			// update
			err := r.UpdateDNSRecord(ctx, cfc, ingress.Hostname, tun, record)
			if err != nil {
				return err
			}
		} else {
			// create
			err := r.CreateDNSRecord(ctx, cfc, ingress.Hostname, zone, tun)
			if err != nil {
				return err
			}
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

// Create a Cloudflare DNS record.
func (r *ArgonautReconciler) CreateDNSRecord(ctx context.Context, cfc *cloudflare.API, name string, zoneid string, tun *cloudflare.ArgoTunnel) error {
	record := cloudflare.DNSRecord{
		Type:      "CNAME",
		Name:      name,
		Content:   tun.ID + ".cfargotunnel.com",
		Proxiable: true,
		Proxied:   new(bool),
		TTL:       1,
		Locked:    false,
		ZoneID:    zoneid,
		Data:      nil,
		Meta:      nil,
		Priority:  nil,
	}
	res, err := cfc.CreateDNSRecord(ctx, zoneid, record)
	if err != nil {
		fmt.Println(res)
		return err
	}
	return nil
}

func (r *ArgonautReconciler) UpdateDNSRecord(ctx context.Context, cfc *cloudflare.API, name string, tun *cloudflare.ArgoTunnel, record cloudflare.DNSRecord) error {
	update := cloudflare.DNSRecord{
		ID:        record.ID,
		Type:      "CNAME",
		Name:      name,
		Content:   tun.ID + ".cfargotunnel.com",
		Proxiable: true,
		Proxied:   record.Proxied,
		TTL:       1,
		Locked:    false,
		ZoneID:    record.ZoneID,
		ZoneName:  record.ZoneName,
		Data:      record.Data,
		Meta:      record.Meta,
		Priority:  record.Priority,
	}

	err := cfc.UpdateDNSRecord(ctx, record.ZoneID, record.ID, update)
	if err != nil {
		return err
	}
	log.FromContext(ctx).Info("Updated DNS Record", "host", name, "cname", record.Content)
	return nil
}

// Checks if a hostname is found in a slice of cloudflare.DNSRecord items.
func inDNSRecords(records []cloudflare.DNSRecord, item string) (bool, cloudflare.DNSRecord) {
	for _, record := range records {
		if record.Name == item {
			return true, record
		}
	}
	return false, cloudflare.DNSRecord{}
}
