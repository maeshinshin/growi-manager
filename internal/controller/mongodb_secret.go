package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"

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

func (r *GrowiReconciler) reconcileMongodbSecret(ctx context.Context, growi *growiv1.Growi) error {
	logger := logf.FromContext(ctx)
	mongodbSecretName := getMongodbSecretName(*growi)

	secret := corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Name: mongodbSecretName, Namespace: growi.Namespace}, &secret)
	if err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get MongoDB secret")
		return err
	} else if err == nil && secret.Data["MONGO_INITDB_ROOT_USERNAME"] != nil && secret.Data["MONGO_INITDB_ROOT_PASSWORD"] != nil && secret.Data["mongo.key"] != nil {
		logger.Info("MongoDB secret already exists, skipping creation")
		return nil
	}

	logger.Info("Creating MongoDB secret")

	mongoSecret := corev1apply.Secret(mongodbSecretName, growi.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       "growi",
			"app.kubernetes.io/instance":   growi.Name,
			"app.kubernetes.io/managed-by": FIELDMANAGER_NAME,
			"app.kubernetes.io/component":  "mongodb",
		}).
		WithType("Opaque").
		WithData(map[string][]byte{
			"MONGO_INITDB_ROOT_USERNAME": generateRandomBytes(ctx, 20),
			"MONGO_INITDB_ROOT_PASSWORD": generateRandomBytes(ctx, 20),
			"mongo.key":                  generateMongodbKeyFileContent(ctx),
		})

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mongoSecret)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	if err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELDMANAGER_NAME,
		Force:        ptr.To(true),
	}); err != nil {
		logger.Error(err, "Failed to create MongoDB secret")
		return err
	}

	return nil
}

func (r *GrowiReconciler) getMongodbUsername(ctx context.Context, growi *growiv1.Growi) string {
	logger := logf.FromContext(ctx)

	var secret corev1.Secret
	if err := r.Get(ctx, client.ObjectKey{Name: getMongodbSecretName(*growi), Namespace: growi.Namespace}, &secret); err != nil {
		logger.Error(err, "Failed to get MongoDB secret")
		return ""
	}
	username := string(secret.Data["MONGO_INITDB_ROOT_USERNAME"])
	return username
}

func (r *GrowiReconciler) getMongodbPassword(ctx context.Context, growi *growiv1.Growi) string {
	logger := logf.FromContext(ctx)

	var secret corev1.Secret
	if err := r.Get(ctx, client.ObjectKey{Name: getMongodbSecretName(*growi), Namespace: growi.Namespace}, &secret); err != nil {
		logger.Error(err, "Failed to get MongoDB secret")
		return ""
	}
	password := string(secret.Data["MONGO_INITDB_ROOT_PASSWORD"])
	return password
}

func generateRandomBytes(ctx context.Context, length int) []byte {
	logger := logf.FromContext(ctx)
	if length <= 0 {
		return nil
	}

	byteSlice, err := generateAlphanumericBytesRecursive(ctx, length)
	if err != nil {
		logger.Error(err, "Failed to generate random bytes")
		return nil
	}

	return byteSlice
}

func generateAlphanumericBytesRecursive(ctx context.Context, length int) ([]byte, error) {
	logger := logf.FromContext(ctx)
	if length <= 0 {
		return nil, nil
	}

	charsetLength := big.NewInt(int64(len(CREDENTIAL_CHARACTER_SET)))

	randomIndex, err := rand.Int(rand.Reader, charsetLength)
	if err != nil {
		logger.Error(err, "Failed to generate random index for alphanumeric bytes")
	}

	restOfBytes, err := generateAlphanumericBytesRecursive(ctx, length-1)
	if err != nil {
		return nil, err
	}

	return append(restOfBytes, CREDENTIAL_CHARACTER_SET[randomIndex.Int64()]), nil
}

func generateMongodbKeyFileContent(ctx context.Context) []byte {
	logger := logf.FromContext(ctx)
	keyLength := 768
	randomBytes := make([]byte, keyLength)

	n, err := rand.Read(randomBytes)
	if err != nil {
		logger.Error(err, "Failed to generate cryptographically secure random bytes for key file")
	}
	if n != keyLength {
		err = fmt.Errorf("failed to generate enough random bytes for key file: expected %d, got %d", keyLength, n) // fmt パッケージが必要
		logger.Error(err, "Incomplete random bytes generation")
	}

	keyFileContent := base64.StdEncoding.EncodeToString(randomBytes)

	logger.Info("Generated random bytes for MongoDB key file")
	return []byte(keyFileContent)
}
