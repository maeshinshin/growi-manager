package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	growiv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func (r GrowiReconciler) reconcileGrowiappSecret(ctx context.Context, growi *growiv1.Growi) error {
	logger := logf.FromContext(ctx)
	growiappSecretName := getGrowiappSecretName(*growi)
	growiappLabels := getLabels(*growi, COMPONENT_GROWIAPP)

	secret := corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Name: growiappSecretName, Namespace: growi.Namespace}, &secret)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get Growiapp secret")
		return err
	} else if err == nil && secret.Data["PASSWORD_SEED"] != nil {
		logger.Info("Growiapp secret already exists, skipping creation")
		return nil
	}

	growiappSecret := corev1apply.Secret(growiappSecretName, growi.Namespace).
		WithLabels(growiappLabels).
		WithType("Opaque").
		WithData(map[string][]byte{
			"PASSWORD_SEED": generateRandomBytes(ctx, 20),
		})

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(growiappSecret)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	logger.Info("Creating Growiapp secret", "name", growiappSecretName)
	if err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELDMANAGER_NAME,
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to create Growiapp secret")
		return err
	}

	return nil
}
