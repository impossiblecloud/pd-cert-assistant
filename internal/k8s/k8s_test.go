package k8s

import (
	"testing"
	"time"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInjectAnnotationsWithNoAnnotation(t *testing.T) {
	// Mock certificate with no annotations
	certificate := cmapi.Certificate{}
	annotations := injectAnnotations(certificate)

	assert.NotNil(t, annotations, "Annotations map should not be nil")
	assert.Equal(t, "pd-assistant", annotations["managed-by"], "Annotation 'managed-by' should be set to 'pd-assistant'")
	assert.NotEmpty(t, annotations["last-updated"], "Annotation 'last-updated' should not be empty")

	// Check if 'last-updated' is a valid timestamp
	_, err := time.Parse("2006-01-02 15:04:05", annotations["last-updated"])
	assert.NoError(t, err, "Annotation 'last-updated' should be a valid timestamp")

	// Mock certificate with existing annotations
	existingAnnotations := map[string]string{
		"existing-key": "existing-value",
	}
	certificate.SetAnnotations(existingAnnotations)
	annotations = injectAnnotations(certificate)

	assert.Equal(t, "pd-assistant", annotations["managed-by"], "Annotation 'managed-by' should be set to 'pd-assistant'")
	assert.NotEmpty(t, annotations["last-updated"], "Annotation 'last-updated' should not be empty")
	assert.Equal(t, "existing-value", annotations["existing-key"], "Existing annotations should be preserved")
}

func TestInjectAnnotationsWithAnnotation(t *testing.T) {
	// Mock certificate with no annotations
	certificate := cmapi.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"existing-key": "existing-value",
			},
		},
	}
	annotations := injectAnnotations(certificate)

	assert.NotNil(t, annotations, "Annotations map should not be nil")
	assert.Equal(t, "pd-assistant", annotations["managed-by"], "Annotation 'managed-by' should be set to 'pd-assistant'")
	assert.NotEmpty(t, annotations["last-updated"], "Annotation 'last-updated' should not be empty")

	// Check if 'existing-key' is preserved
	assert.Equal(t, "existing-value", annotations["existing-key"], "Existing annotations should be preserved")

	// Check if 'last-updated' is a valid timestamp
	_, err := time.Parse("2006-01-02 15:04:05", annotations["last-updated"])
	assert.NoError(t, err, "Annotation 'last-updated' should be a valid timestamp")

	// Mock certificate with existing annotations
	existingAnnotations := map[string]string{
		"existing-key": "existing-value",
	}
	certificate.SetAnnotations(existingAnnotations)
	annotations = injectAnnotations(certificate)

	assert.Equal(t, "pd-assistant", annotations["managed-by"], "Annotation 'managed-by' should be set to 'pd-assistant'")
	assert.NotEmpty(t, annotations["last-updated"], "Annotation 'last-updated' should not be empty")
	assert.Equal(t, "existing-value", annotations["existing-key"], "Existing annotations should be preserved")
}
