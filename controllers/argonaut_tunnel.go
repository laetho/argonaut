package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/ghodss/yaml"
	argonautv1 "github.com/laetho/argonaut/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
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
	// Create ConfigMap, will be mapped into Pod
	if err := r.ReconcileArgonautTunnelSecret(ctx, argonaut, &tun, cfc.AccountID); err != nil {
		return &cloudflare.ArgoTunnel{}, err
	}
	// Create Secret, will be mapped into Pod
	if err := r.ReconcileArgonautTunnelConfig(ctx, argonaut, &tun); err != nil {
		return &cloudflare.ArgoTunnel{}, err
	}

	return &tun, nil
}

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
func (r *ArgonautReconciler) ReconcileArgonautTunnelSecret(ctx context.Context, argonaut *argonautv1.Argonaut, tun *cloudflare.ArgoTunnel, account string) error {
	var secret v1.Secret

	payload, err := json.Marshal(ArgonautTunnelSecret{
		AccountTag:   account,
		TunnelSecret: tun.Secret,
		TunnelID:     tun.ID,
		TunnelName:   tun.Name,
	})
	if err != nil {
		return err
	}

	if err := r.Get(ctx, client.ObjectKey{Name: argonaut.Name, Namespace: argonaut.Namespace}, &secret); err != nil {
		log.FromContext(ctx).Info("Argonaut tunnel secret not found, creating", "secret", argonaut.Name)
		secret.Name = argonaut.Name
		secret.Namespace = argonaut.Namespace
		secret.StringData = make(map[string]string)
		secret.StringData["tunnel.json"] = string(payload)

		if err := r.Create(ctx, &secret); err != nil {
			return err
		}
	} else {
		secret.Name = argonaut.Name
		secret.Namespace = argonaut.Namespace
		secret.StringData = make(map[string]string)
		secret.StringData["tunnel.json"] = string(payload)

		if err := r.Update(ctx, &secret); err != nil {
			return err
		}
		log.FromContext(ctx).Info("Updated Secret", "name", secret.Name)
	}
	return nil
}

// Creates or updates a ConfigMap with the ArgoTunnel configuration.
func (r *ArgonautReconciler) ReconcileArgonautTunnelConfig(ctx context.Context, argonaut *argonautv1.Argonaut, tun *cloudflare.ArgoTunnel) error {
	var conf v1.ConfigMap

	payload, err := yaml.Marshal(r.BuildArgonautTunnelConfig(ctx, argonaut, tun))
	if err != nil {
		return err
	}

	if err := r.Get(ctx, client.ObjectKey{Name: argonaut.Name, Namespace: argonaut.Namespace}, &conf); err != nil {
		log.FromContext(ctx).Info("Did not find ConfigMap, creating", "name", argonaut.Name)
		conf.Name = argonaut.Name
		conf.Namespace = argonaut.Namespace
		conf.Data = make(map[string]string)
		conf.Data["config.yml"] = string(payload)

		if err := r.Create(ctx, &conf); err != nil {
			return err
		}
	} else {
		conf.Name = argonaut.Name
		conf.Namespace = argonaut.Namespace
		conf.Data = make(map[string]string)
		conf.Data["config.yml"] = string(payload)

		if err := r.Update(ctx, &conf); err != nil {
			return err
		}
		log.FromContext(ctx).Info("Updated ConfigMap", "name", conf.Name)
	}
	return nil
}

// Deletes an Argo Tunnel using the Cloudflare API
func (r *ArgonautReconciler) DeleteArgoTunnel(ctx context.Context, cfc *cloudflare.API, tun *cloudflare.ArgoTunnel) error {
	if err := cfc.DeleteArgoTunnel(ctx, cfc.AccountID, tun.ID); err != nil {
		return err
	}
	log.FromContext(ctx).Info("deleted Argo Tunnel", tun.ID, tun.Name)
	return nil
}

// Builds the cloudflared config.yml from an Argonaut objekt with endpoint selectors etc.
func (r *ArgonautReconciler) BuildArgonautTunnelConfig(ctx context.Context, argonaut *argonautv1.Argonaut, tun *cloudflare.ArgoTunnel) ArgonautTunnelConfig {
	conf := ArgonautTunnelConfig{
		Tunnel:          tun.ID,
		CredentialsFile: "/etc/cloudflared/tunnel.json",
		Ingress:         nil,
	}
	var ingressConf []ArgonautTunnelConfigIngress

	for _, ingress := range argonaut.Spec.Ingress {
		var svc v1.ServiceList
		err := r.List(ctx, &svc, client.MatchingLabels(ingress.ServiceSelector.MatchLabels))
		if err != nil {
			log.FromContext(ctx).Info("Found no Service matching selector", "selector", ingress.ServiceSelector)
		}

		// Find ClusterIP and Ports for each Service and create a ArgonautTunnelConfigIngress
		// This one is very naive, and should be changed
		for _, service := range svc.Items {
			clusterip := service.Spec.ClusterIP
			port := strconv.Itoa(int(service.Spec.Ports[0].Port))
			protocol := "http://"
			ingressConf = append(ingressConf, ArgonautTunnelConfigIngress{
				Hostname: ingress.Hostname,
				Service:  protocol + clusterip + ":" + port,
			})
		}
	}

	ingressConf = append(ingressConf, ArgonautTunnelConfigIngress{
		Service: "http_status:404",
	})
	conf.Ingress = ingressConf

	return conf
}
