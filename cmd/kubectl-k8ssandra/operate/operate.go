package operate

import (
	"fmt"

	"github.com/k8ssandra/k8ssandra-client/pkg/cassdcutil"
	"github.com/k8ssandra/k8ssandra-client/pkg/kubernetes"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	startExample = `
	# start an existing datacenter that was stopped
	%[1]s start <datacenter>
	`

	stopExample = `
	# shutdown an existing datacenter
	%[1]s stop <datacenter>

	# shutdown an existing datacenter and wait for all the pods to shutdown
	%[1]s stop <datacenter> --wait
	`

	restartExample = `
	# request a rolling restart for datacenter
	%[1]s restart <datacenter>

	# request a rolling restart of a single rack called r1
	%[1]s restart <datacenter> --rack r1
	`

	errNoDatacenterDefined = fmt.Errorf("no target datacenter given")
	errRestartingStopped   = fmt.Errorf("unable to do rolling restart to a stopped datacenter")
)

type options struct {
	configFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams
	namespace   string
	dcName      string
	rackName    string
	wait        bool
	cassManager *cassdcutil.CassManager
}

func newOptions(streams genericclioptions.IOStreams) *options {
	return &options{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

func NewStartCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := newOptions(streams)

	cmd := &cobra.Command{
		Use:          "start [cluster]",
		Short:        "restart an existing shutdown Cassandra cluster",
		Example:      fmt.Sprintf(startExample, "kubectl k8ssandra"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(false); err != nil {
				return err
			}

			return nil
		},
	}

	fl := cmd.Flags()
	fl.BoolVarP(&o.wait, "wait", "w", false, "wait until all pods have started")
	fl.StringVar(&o.rackName, "rack", "", "restart only target rack")
	o.configFlags.AddFlags(fl)
	return cmd
}

func NewRestartCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := newOptions(streams)

	cmd := &cobra.Command{
		Use:          "restart [cluster]",
		Short:        "request rolling restart for an existing running Cassandra cluster",
		Example:      fmt.Sprintf(restartExample, "kubectl k8ssandra"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.ValidateRestart(); err != nil {
				return err
			}
			if err := o.Restart(); err != nil {
				return err
			}

			return nil
		},
	}

	fl := cmd.Flags()
	fl.BoolVarP(&o.wait, "wait", "w", false, "wait until all pods have restarted")
	o.configFlags.AddFlags(fl)
	return cmd
}

func NewStopCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := newOptions(streams)

	cmd := &cobra.Command{
		Use:          "stop [cluster]",
		Short:        "shutdown running Cassandra cluster",
		Example:      fmt.Sprintf(stopExample, "kubectl k8ssandra"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(true); err != nil {
				return err
			}

			return nil
		},
	}

	fl := cmd.Flags()
	fl.BoolVarP(&o.wait, "wait", "w", false, "wait until all pods have terminated")
	o.configFlags.AddFlags(fl)
	return cmd
}

// Complete parses the arguments and necessary flags to options
func (c *options) Complete(cmd *cobra.Command, args []string) error {
	var err error

	if len(args) < 1 {
		return errNoDatacenterDefined
	}

	c.dcName = args[0]

	c.namespace, _, err = c.configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	restConfig, err := c.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	kubeClient, err := kubernetes.GetClientInNamespace(restConfig, c.namespace)
	if err != nil {
		return err
	}

	c.cassManager = cassdcutil.NewManager(kubeClient)

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (c *options) Validate() error {
	// Verify target cluster exists
	_, err := c.cassManager.CassandraDatacenter(c.dcName, c.namespace)
	if err != nil {
		// NotFound is still an error
		return err
	}
	return nil
}

// ValidateRestart ensures that all required arguments and flag values are provided
func (c *options) ValidateRestart() error {
	// Verify target cluster exists
	dc, err := c.cassManager.CassandraDatacenter(c.dcName, c.namespace)
	if err != nil {
		// NotFound is still an error
		return err
	}
	if dc.Spec.Stopped {
		return errRestartingStopped
	}
	return nil
}

// Run either stops or starts the existing datacenter
func (c *options) Run(stop bool) error {
	return c.cassManager.ModifyStoppedState(c.dcName, c.namespace, stop, c.wait)
}

// Restart creates a restart task for the cluster
func (c *options) Restart() error {
	return c.cassManager.RestartDc(c.dcName, c.namespace, c.rackName, c.wait)
}
