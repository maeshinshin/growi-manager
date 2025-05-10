package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"math/big"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	growiv1 "github.com/maeshinshin/growi-manager/api/v1"
)

func (r *GrowiReconciler) reconcileMongoDBSecret(ctx context.Context, growi *growiv1.Growi) error {
	logger := log.FromContext(ctx)

	secret := corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Name: getMongoDBSecretName(*growi), Namespace: growi.Namespace}, &secret)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get MongoDB secret")
		return err
	} else if err == nil && secret.Data["MONGO_INITDB_ROOT_USERNAME"] != nil && secret.Data["MONGO_INITDB_ROOT_PASSWORD"] != nil {
		logger.Info("MongoDB secret already exists, skipping creation")
		if *growi.Status.MongoDBSecretStatus != growiv1.ExistMongoDBSecret {
			growi.Status.MongoDBSecretStatus = ptr.To(growiv1.ExistMongoDBSecret)
			if err := r.Status().Update(ctx, growi); err != nil {
				logger.Error(err, "Failed to update Growi status")
				return err
			}
		}
		return nil
	}

	logger.Info("Creating MongoDB secret")
	growi.Status.MongoDBSecretStatus = ptr.To(growiv1.CreatingMongoDBSecret)
	if err := r.Status().Update(ctx, growi); err != nil {
		logger.Error(err, "Failed to update Growi status")
		return err
	}

	mongoSecret := corev1apply.Secret(getMongoDBSecretName(*growi), growi.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       "growi",
			"app.kubernetes.io/instance":   growi.Name,
			"app.kubernetes.io/managed-by": "growi-manager",
		}).
		WithType("Opaque").
		WithData(map[string][]byte{
			"MONGO_INITDB_ROOT_USERNAME": generateRandomBytes(ctx),
			"MONGO_INITDB_ROOT_PASSWORD": generateRandomBytes(ctx),
		})

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mongoSecret)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	if err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "growi-manager",
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to create MongoDB secret")
		growi.Status.MongoDBSecretStatus = ptr.To(growiv1.FailedtoCreateMongoDBSecret)
		if err := r.Status().Update(ctx, growi); err != nil {
			logger.Error(err, "Failed to update Growi status")
		}
		return err
	}

	growi.Status.MongoDBSecretStatus = ptr.To(growiv1.ExistMongoDBSecret)
	if err := r.Status().Update(ctx, growi); err != nil {
		logger.Error(err, "Failed to update Growi status")
	}

	return nil
}

func (r *GrowiReconciler) getMongoDBUsernameAndPassword(ctx context.Context, growi *growiv1.Growi) (string, string) {
	logger := log.FromContext(ctx)

	var secret corev1.Secret
	if err := r.Get(ctx, client.ObjectKey{Name: getMongoDBSecretName(*growi), Namespace: growi.Namespace}, &secret); err != nil {
		logger.Error(err, "Failed to get MongoDB secret")
		return "", ""
	}
	username := string(secret.Data["MONGO_INITDB_ROOT_USERNAME"])
	password := string(secret.Data["MONGO_INITDB_ROOT_PASSWORD"])
	return username, password
}

func generateRandomBytes(ctx context.Context) []byte {
	logger := log.FromContext(ctx)
	randomLength, err := rand.Int(rand.Reader, big.NewInt(6))
	if err != nil {
		logger.Error(err, "Failed to generate random length")
	}

	bytes := make([]byte, 15+randomLength.Int64())
	_, err = rand.Read(bytes)
	if err != nil {
		logger.Error(err, "Failed to generate random bytes")
	}
	logger.Info("Generated random bytes")
	return []byte(base64.StdEncoding.EncodeToString(bytes))
}
