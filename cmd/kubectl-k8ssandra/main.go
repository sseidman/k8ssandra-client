package main

import (
	"os"

	"github.com/spf13/pflag"

	"github.com/k8ssandra/k8ssandra-client/cmd/kubectl-k8ssandra/k8ssandra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-k8ssandra", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := k8ssandra.NewCmd(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
