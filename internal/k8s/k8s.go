package k8s

import (
	"context"
	"fmt"
	"time"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmclient "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	"github.com/golang/glog"
	"github.com/impossiblecloud/pd-cert-assistant/internal/cfg"
	"github.com/impossiblecloud/pd-cert-assistant/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Config *rest.Config
}

func loadKubeConfig(path string) (*rest.Config, error) {
	file, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	// Just in case we decide to do some overrides in the future
	override := &clientcmd.ConfigOverrides{}
	return clientcmd.NewDefaultClientConfig(*file, override).ClientConfig()
}

func injectAnnotations(certificate cmapi.Certificate) map[string]string {
	annotations := certificate.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["managed-by"] = "pd-assistant"
	annotations["last-updated"] = time.Now().UTC().Format(time.RFC3339)
	return annotations
}

func (c *Client) Init(kubeConfigPath string) error {
	var err error

	// Load either the kubeconfig file or in-cluster config
	if kubeConfigPath != "" {
		c.Config, err = loadKubeConfig(kubeConfigPath)
	} else {
		c.Config, err = rest.InClusterConfig()
	}
	return err
}

// GetCiliumNodes retrieves a list of CiliumNode resources from the cilium.io/v2 API
// and returns a list of CiliumInternalIP addresses.
func (c *Client) GetCiliumNodes() ([]string, error) {

	ciliumNodeGVR := schema.GroupVersionResource{
		Group:    "cilium.io",
		Version:  "v2",
		Resource: "ciliumnodes",
	}

	dynamicClient, err := dynamic.NewForConfig(c.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	ciliumNodes, err := dynamicClient.Resource(ciliumNodeGVR).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CiliumNode resources: %v", err)
	}

	// Extract CiliumInternalIP addresses
	var internalIPs []string
	for _, item := range ciliumNodes.Items {
		glog.V(8).Infof("Processing CiliumNode: %s", item.GetName())
		spec, ok := item.Object["spec"].(map[string]interface{})
		if !ok {
			continue
		}

		addresses, ok := spec["addresses"].([]interface{})
		if !ok {
			continue
		}

		for _, addr := range addresses {
			addressMap, ok := addr.(map[string]interface{})
			if !ok {
				continue
			}

			if addressMap["type"] == "CiliumInternalIP" {
				if ip, ok := addressMap["ip"].(string); ok {
					internalIPs = append(internalIPs, ip)
				}
			}
		}
	}

	return internalIPs, nil
}

// UpdateCertificate updates the certificate in Kubernetes with the provided IP addresses.
func (c *Client) UpdateCertificate(conf cfg.AppConfig, IPs []string) error {
	client, err := cmclient.NewForConfig(c.Config)
	if err != nil {
		return err
	}

	// Override IP addresses from the configuration
	conf.Certificate.Spec.IPAddresses = IPs
	conf.Certificate.SetAnnotations(injectAnnotations(conf.Certificate))

	certificate, err := client.CertmanagerV1().Certificates(conf.Certificate.Namespace).Get(context.TODO(), conf.Certificate.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			glog.Infof("Certificate %s/%s not found, creating a new one", conf.Certificate.Namespace, conf.Certificate.Name)
			_, err = client.CertmanagerV1().Certificates(conf.Certificate.Namespace).Create(context.TODO(), &conf.Certificate, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create certificate %s/%s: %s", conf.Certificate.Namespace, conf.Certificate.Name, err.Error())
			}
			glog.Infof("Certificate %s/%s created successfully", conf.Certificate.Namespace, conf.Certificate.Name)
			return nil
		}
		return fmt.Errorf("failed to get certificate %s/%s: %s", certificate.Namespace, conf.Certificate.Name, err.Error())
	}

	// Check if the IPs are already set and are the same as the current ones
	if utils.IPListsEqual(certificate.Spec.IPAddresses, IPs) {
		glog.V(6).Infof("Certificate %s/%s already has the same IPs, no update needed", conf.Certificate.Namespace, conf.Certificate.Name)
		return nil
	}

	// Update Certificate IPs and some annotations
	glog.V(6).Infof("Certificate %s/%s found, updating IPs: %v", conf.Certificate.Namespace, conf.Certificate.Name, IPs)
	certificate.Spec.IPAddresses = IPs
	certificate.SetAnnotations(injectAnnotations(*certificate))
	_, err = client.CertmanagerV1().Certificates(conf.Certificate.Namespace).Update(context.TODO(), certificate, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update certificate %s/%s: %s", conf.Certificate.Namespace, conf.Certificate.Name, err.Error())
	}
	glog.Infof("Certificate %s/%s updated successfully", conf.Certificate.Namespace, conf.Certificate.Name)
	return nil
}
