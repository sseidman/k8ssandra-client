package secrets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/k8ssandra/k8ssandra-client/pkg/kubernetes"
	"github.com/k8ssandra/k8ssandra-client/pkg/secrets"
	"github.com/k8ssandra/k8ssandra-operator/controllers/secrets-webhook"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	mountExample = `
        # mount secrets to the local file system
        %[1]s mount <injections> <namespace>
        `

	errNoInjectionsDefined = fmt.Errorf("no injection string provided")
)

type options struct {
	configFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams
	injections string
	namespace  string
	client     kubernetes.NamespacedClient
}

func newOptions(streams genericclioptions.IOStreams) *options {
	return &options{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

func NewMountCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := newOptions(streams)

	cmd := &cobra.Command{
		Use:          "mount [injections]",
		Short:        "mount secrets to the local file system",
		Example:      fmt.Sprintf(mountExample, "kubectl k8ssandra"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&o.injections, "injections", "i", "", "secrets to inject into filesystem")
	o.configFlags.AddFlags(flags)
	return cmd
}

func (c *options) Complete(cmd *cobra.Command, args []string) error {
	var err error

	c.namespace, _, err = c.configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	restConfig, err := c.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	c.client, err = kubernetes.GetClientInNamespace(restConfig, c.namespace)
	return err
}

func (c *options) Validate(args []string) error {
	if len(args) < 1 {
		return errNoInjectionsDefined
	}

	c.injections = args[0]
	return nil
}

func (c *options) Run() error {
	var si []secrets_webhook.SecretInjection
	if err := json.Unmarshal([]byte(c.injections), &si); err != nil {
		return err
	}

	for _, injection := range si {
		secret := &corev1.Secret{}
		err := c.client.Get(context.Background(), types.NamespacedName{Name: injection.SecretName}, secret)
		if err != nil {
			return err
		}

		err = secrets.CreateSecretsDirectory(injection.Path, injection.SecretName)
		if err != nil {
			return err
		}

		for k, v := range secret.Data {
			err = secrets.WriteSecretsKeyValue(injection.Path, injection.SecretName, k, string(v))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
