package resources

import (
	"context"
	"os"
	"testing"

	"github.com/stackrox/rox/generated/internalapi/central"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/registries/types"
	"github.com/stackrox/rox/sensor/common/registry"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	openshift311DockerConfigSecret = &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-dockercfg-6167c",
			Namespace: "test-ns",
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": "default",
			},
		},
		Data: map[string][]byte{
			".dockercfg": []byte(`
{
  "docker-registry.default.svc.cluster.local:5000": {
    "username": "serviceaccount",
    "password": "password",
    "email": "serviceaccount@example.org"
  }
}`),
		},
		Type: "kubernetes.io/dockercfg",
	}

	openshift4xDockerConfigSecret = &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-dockercfg-9w5gn",
			Namespace: "test-ns",
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": "default",
			},
		},
		Data: map[string][]byte{
			".dockercfg": []byte(`
{
  "image-registry.openshift-image-registry.svc:5000": {
    "username": "serviceaccount",
    "password": "password",
    "email": "serviceaccount@example.org"
  }
}`),
		},
		Type: "kubernetes.io/dockercfg",
	}
)

// alwaysInsecureCheckTLS is an implementation of registry.CheckTLS
// which always says the given address is insecure.
func alwaysInsecureCheckTLS(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func TestOpenShiftRegistrySecret_311(t *testing.T) {
	regStore := registry.NewRegistryStore(alwaysInsecureCheckTLS)
	d := newSecretDispatcher(regStore)

	_ = d.ProcessEvent(openshift311DockerConfigSecret, nil, central.ResourceAction_CREATE_RESOURCE)

	imgName := &storage.ImageName{
		Registry: "docker-registry.default.svc.cluster.local:5000",
		Remote:   "dummy/nginx",
		Tag:      "1.18.0",
		FullName: "docker-registry.default.svc.cluster.local:5000/stackrox/nginx:1.18.0",
	}

	reg, err := regStore.GetRegistryForImage(imgName)
	assert.Nil(t, reg)
	assert.Error(t, err)

	imgName = &storage.ImageName{
		Registry: "docker-registry.default.svc.cluster.local:5000",
		Remote:   "test-ns/nginx",
		Tag:      "1.18.0",
		FullName: "docker-registry.default.svc.cluster.local:5000/stackrox/nginx:1.18.0",
	}

	reg, err = regStore.GetRegistryForImage(imgName)
	assert.NotNil(t, reg)
	assert.NoError(t, err)

	expectedRegConfig := &types.Config{
		Username:         "serviceaccount",
		Password:         "password",
		Insecure:         true,
		URL:              "https://docker-registry.default.svc.cluster.local:5000",
		RegistryHostname: "docker-registry.default.svc.cluster.local:5000",
		Autogenerated:    false,
	}

	assert.Equal(t, "docker-registry.default.svc.cluster.local:5000", reg.Name())
	assert.Equal(t, expectedRegConfig, reg.Config())
}

func TestOpenShiftRegistrySecret_4x(t *testing.T) {
	regStore := registry.NewRegistryStore(alwaysInsecureCheckTLS)
	d := newSecretDispatcher(regStore)

	_ = d.ProcessEvent(openshift4xDockerConfigSecret, nil, central.ResourceAction_CREATE_RESOURCE)

	imgName := &storage.ImageName{
		Registry: "image-registry.openshift-image-registry.svc:5000",
		Remote:   "dummy/nginx",
		Tag:      "1.18.0",
		FullName: "image-registry.openshift-image-registry.svc:5000/stackrox/nginx:1.18.0",
	}

	reg, err := regStore.GetRegistryForImage(imgName)
	assert.Nil(t, reg)
	assert.Error(t, err)

	imgName = &storage.ImageName{
		Registry: "image-registry.openshift-image-registry.svc:5000",
		Remote:   "test-ns/nginx",
		Tag:      "1.18.0",
		FullName: "image-registry.openshift-image-registry.svc:5000/stackrox/nginx:1.18.0",
	}

	reg, err = regStore.GetRegistryForImage(imgName)
	assert.NotNil(t, reg)
	assert.NoError(t, err)

	expectedRegConfig := &types.Config{
		Username:         "serviceaccount",
		Password:         "password",
		Insecure:         true,
		URL:              "https://image-registry.openshift-image-registry.svc:5000",
		RegistryHostname: "image-registry.openshift-image-registry.svc:5000",
		Autogenerated:    false,
	}

	assert.Equal(t, "image-registry.openshift-image-registry.svc:5000", reg.Name())
	assert.Equal(t, expectedRegConfig, reg.Config())
}

// TestProcessDockerConfig_ForceLocalScanning tests that dockerconfig secrets are stored
// in the regStore as expected when local scanning is forced
func TestProcessDockerConfig_ForceLocalScanning(t *testing.T) {
	fakeNamespace := "fake-namespace"
	dockerConfigSecret := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "fake-secret", Namespace: fakeNamespace},
		Type:       v1.SecretTypeDockercfg,
		Data: map[string][]byte{v1.DockerConfigKey: []byte(`
			{
				"fake.reg.local": {
					"username": "hello",
					"password": "world",
					"email": "hello@example.com"
				}
			}
		`)},
	}
	fakeImage := &storage.ImageName{
		Registry: "fake.reg.local",
		Remote:   "fake/repo",
		Tag:      "latest",
		FullName: "fake.reg.local/fake/repo:latest",
	}

	origEnv := os.Getenv("ROX_FORCE_LOCAL_IMAGE_SCANNING")
	defer os.Setenv("ROX_FORCE_LOCAL_IMAGE_SCANNING", origEnv)

	os.Setenv("ROX_FORCE_LOCAL_IMAGE_SCANNING", "false")

	// with feature disabled, registry secret should NOT be stored
	regStore := registry.NewRegistryStore(alwaysInsecureCheckTLS)
	d := newSecretDispatcher(regStore)

	d.ProcessEvent(dockerConfigSecret, nil, central.ResourceAction_CREATE_RESOURCE)
	reg, err := regStore.GetRegistryForImageInNamespace(fakeImage, fakeNamespace)
	assert.Nil(t, reg)
	assert.Error(t, err)

	os.Setenv("ROX_FORCE_LOCAL_IMAGE_SCANNING", "true")

	// feature is enabled, registry secret should be stored
	d.ProcessEvent(dockerConfigSecret, nil, central.ResourceAction_CREATE_RESOURCE)
	reg, err = regStore.GetRegistryForImageInNamespace(fakeImage, fakeNamespace)
	assert.NotNil(t, reg)
	assert.NoError(t, err)
	assert.Equal(t, reg.Config().Username, "hello")

	regStore = registry.NewRegistryStore(alwaysInsecureCheckTLS)
	d = newSecretDispatcher(regStore)

	// secrets with an service-account.name other than default should not be stored
	dockerConfigSecret.Annotations = map[string]string{saAnnotation: "something"}

	d.ProcessEvent(dockerConfigSecret, nil, central.ResourceAction_CREATE_RESOURCE)
	reg, err = regStore.GetRegistryForImageInNamespace(fakeImage, fakeNamespace)
	assert.Nil(t, reg)
	assert.Error(t, err)

	// secrets with an saAnnotation of `default` should still be stored
	dockerConfigSecret.Annotations = map[string]string{saAnnotation: "default"}

	d.ProcessEvent(dockerConfigSecret, nil, central.ResourceAction_CREATE_RESOURCE)
	reg, err = regStore.GetRegistryForImageInNamespace(fakeImage, fakeNamespace)
	assert.NotNil(t, reg)
	assert.NoError(t, err)
}
