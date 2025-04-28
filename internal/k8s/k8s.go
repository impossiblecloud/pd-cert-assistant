package k8s

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/impossiblecloud/pd-cert-assistant/internal/cfg"
	"github.com/impossiblecloud/pd-cert-assistant/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	cmclient "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
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

func (c *Client) UpdateCertificate(conf cfg.AppConfig, IPs []string) error {
	client, err := cmclient.NewForConfig(c.Config)
	if err != nil {
		return err
	}
	certificate, err := client.CertmanagerV1().Certificates(conf.CertificateNamespace).Get(context.TODO(), conf.CertificateName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			glog.Info("Certificate %s/%s not found, creating a new one", conf.CertificateNamespace, conf.CertificateName)
			// TODO: read certificate from YAML file and then override the IPs
			newCert := &cmapi.Certificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      conf.CertificateName,
					Namespace: conf.CertificateNamespace,
				},
				Spec: cmapi.CertificateSpec{
					SecretName: conf.CertificateName,
					IssuerRef: cmmeta.ObjectReference{
						Name: "tidb-cluster-selfsigned-ca-issuer",
						Kind: "Issuer",
					},
					IPAddresses: IPs,
				},
			}
			_, err = client.CertmanagerV1().Certificates(conf.CertificateNamespace).Create(context.TODO(), newCert, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create certificate %s/%s: %s", conf.CertificateNamespace, conf.CertificateName, err.Error())
			}
			glog.Infof("Certificate %s/%s created successfully", conf.CertificateNamespace, conf.CertificateName)
			return nil
		}
		return fmt.Errorf("failed to get certificate %s/%s: %s", certificate.Namespace, conf.CertificateName, err.Error())
	}

	// Check if the IPs are already set and are the same as the current ones
	if utils.IPListsEqual(certificate.Spec.IPAddresses, IPs) {
		glog.V(6).Infof("Certificate %s/%s already has the same IPs, no update needed", conf.CertificateNamespace, conf.CertificateName)
		return nil
	}

	glog.V(6).Infof("Certificate %s/%s found, updating IPs: %v", conf.CertificateNamespace, conf.CertificateName, IPs)
	certificate.Spec.IPAddresses = IPs
	_, err = client.CertmanagerV1().Certificates(conf.CertificateNamespace).Update(context.TODO(), certificate, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update certificate %s/%s: %s", conf.CertificateNamespace, conf.CertificateName, err.Error())
	}
	glog.Infof("Certificate %s/%s updated successfully", conf.CertificateNamespace, conf.CertificateName)
	return nil
}
