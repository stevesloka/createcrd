package main

import (
	"context"
	"flag"
	"log"
	"path/filepath"

	contour "github.com/projectcontour/contour/apis/projectcontour/v1"
	apiv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := apiv1client.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	proxyCRD := &apiv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "httpproxies.projectcontour.io",
		},
		Spec: apiv1.CustomResourceDefinitionSpec{
			Group: contour.GroupName,
			Versions: []apiv1.CustomResourceDefinitionVersion{{
				Name:    "v1",
				Served:  true,
				Storage: true,
				Schema: &apiv1.CustomResourceValidation{
					OpenAPIV3Schema: &apiv1.JSONSchemaProps{
						Type:        "object",
						Description: "HTTPProxy is an Ingress CRD specification",
						Properties: map[string]apiv1.JSONSchemaProps{
							"spec": {
								Type: "object",
								Properties: map[string]apiv1.JSONSchemaProps{
									"virtualhost": {
										Type: "string",
									},
								},
							},
						},
					},
				},
			}},
			Names: apiv1.CustomResourceDefinitionNames{
				Plural:     "httpproxies",
				Singular:   "httpproxy",
				Kind:       "HTTPProxy",
				ListKind:   "HTTPProxList",
				ShortNames: []string{"proxy"},
			},
			Scope:                 apiv1.NamespaceScoped,
			PreserveUnknownFields: false,
		},
	}

	log.Print("Registering HTTPPRoxy CRD")
	_, err = clientset.CustomResourceDefinitions().Create(context.TODO(), proxyCRD, v1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Print("HTTPPRoxy CRD already registered")
		} else {
			errExit("Failed to create HTTPProxy CRD", err)
		}
	}
}

func errExit(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %#v", msg, err)
	}
}
