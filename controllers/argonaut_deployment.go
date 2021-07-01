package controllers

import (
	"context"
	argonautv1 "github.com/laetho/argonaut/api/v1beta1"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciles a Deployment for an Argonaut instance. This is a deployment of the
// cloudflare/cloudflared container with config and secrets.
func (r *ArgonautReconciler) ReconcileArgonautDeployment(ctx context.Context, argonaut *argonautv1.Argonaut) error {

	ownerRef := metav1.OwnerReference{
		APIVersion:         argonaut.APIVersion,
		Kind:               argonaut.Kind,
		Name:               argonaut.Name,
		UID:                argonaut.UID,
		Controller:         nil,
		BlockOwnerDeletion: nil,
	}

	labels := make(map[string]string)
	labels["argonaut"] = argonaut.Name

	labelSelector := metav1.LabelSelector{
		MatchLabels: labels,
	}

	tunnelSecretVolume := v12.Volume{
		Name: "tunnelsecret",
		VolumeSource: v12.VolumeSource{
			Secret: &v12.SecretVolumeSource{
				SecretName: argonaut.Spec.ArgoTunnelName,
			},
		},
	}

	tunnelConfigVolume := v12.Volume{
		Name: "tunnelconfig",
		VolumeSource: v12.VolumeSource{
			ConfigMap: &v12.ConfigMapVolumeSource{
				LocalObjectReference: v12.LocalObjectReference{
					Name: argonaut.Name,
				},
			},
		},
	}

	tunnelSecretVolumeMount := v12.VolumeMount{
		Name:      "tunnelsecret",
		ReadOnly:  true,
		MountPath: "/etc/cloudflare/tunnels",
	}

	tunnelSecretConfigMount := v12.VolumeMount{
		Name:      "tunnelconfig",
		ReadOnly:  true,
		MountPath: "/etc/cloudflare/config",
	}

	containerTemplate := v12.Container{
		Name:                     "cloudflared",
		Image:                    "cloudflare/cloudflared:2021.6.0",
		Command:                  []string{"cloudflared"},
		Args:                     []string{"tunnel", "--config", "/etc/cloudflare/config/config.yaml", "run"},
		VolumeMounts:             append([]v12.VolumeMount{}, tunnelSecretVolumeMount, tunnelSecretConfigMount),
		LivenessProbe:            nil,
		ReadinessProbe:           nil,
		StartupProbe:             nil,
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: "File",
		ImagePullPolicy:          "IfNotPresent",
		Stdin:                    false,
		StdinOnce:                false,
		TTY:                      true,
	}

	var deployment v1.Deployment
	if err := r.Get(ctx, client.ObjectKey{Name: argonaut.Name, Namespace: argonaut.Namespace}, &deployment); err != nil {
		// Create Deployment
		var replicas int32 = 1

		deployment.Name = argonaut.Name
		deployment.Namespace = argonaut.Namespace
		deployment.ObjectMeta.Labels = labels
		deployment.OwnerReferences = append(deployment.OwnerReferences, ownerRef)
		deployment.Spec.Selector = &labelSelector
		deployment.Spec.Replicas = &replicas
		deployment.Spec.Template.Name = argonaut.Name
		deployment.Spec.Template.Labels = labels
		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, tunnelSecretVolume)
		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, tunnelConfigVolume)
		deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, containerTemplate)

		//out, _ := yaml.Marshal(deployment)
		//fmt.Println(string(out))

		if err := r.Create(ctx, &deployment); err != nil {
			return err
		}
		log.FromContext(ctx).Info("Created Argonaut Deployment", "name", deployment.Name)

	} else {
		// Update Deployment
		deployment.Name = argonaut.Name
		deployment.Namespace = argonaut.Namespace
		deployment.ObjectMeta.Labels = labels
		deployment.OwnerReferences = append([]metav1.OwnerReference{}, ownerRef)
		deployment.Spec.Selector = &labelSelector
		deployment.Spec.Template.Name = argonaut.Name
		deployment.Spec.Template.Labels = labels
		deployment.Spec.Template.Spec.Volumes = append([]v12.Volume{}, tunnelSecretVolume, tunnelConfigVolume)
		deployment.Spec.Template.Spec.Containers = append([]v12.Container{}, containerTemplate)

		if err := r.Update(ctx, &deployment); err != nil {
			return err
		}
		log.FromContext(ctx).Info("Updated Argonaut Deployment", "name", deployment.Name)
	}

	return nil
}

func (r *ArgonautReconciler) BuildDeployment(ctx context.Context) (v1.Deployment, error) {
	var deploy v1.Deployment

	return deploy, nil
}
