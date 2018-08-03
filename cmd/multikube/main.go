package main

import (
	"os"
	"fmt"
	"log"
	"github.com/spf13/pflag"
	"gitlab.com/amimof/multikube"
	"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/tools/clientcmd/api"
)

var kubeconfigPath string

func init() {
		pflag.StringVar(&kubeconfigPath, "kubeconfig", "~/.kube/config", "Absolute path to the kubeconfig file")
}

func main() {

	pflag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage:\n")
		fmt.Fprint(os.Stderr, "  multikube-server [OPTIONS]\n\n")

		title := "Kubernetes multi-cluster manager"
		fmt.Fprint(os.Stderr, title+"\n\n")
		desc := "Manages multiple Kubernetes clusters and provides a single API to clients"
		if desc != "" {
			fmt.Fprintf(os.Stderr, desc+"\n\n")
		}
		fmt.Fprintln(os.Stderr, pflag.CommandLine.FlagUsages())
	}

	// parse the CLI flags
	pflag.Parse()

	// Read provided kubeconfig file
	c, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create the proxy
	p := multikube.NewProxyFrom(c)
	m := p.Use(
		multikube.WithEmpty,
		multikube.WithLogging,
		multikube.WithValidate,
	)

	// Create the server
	s := multikube.NewServer(m(p))

	err = s.Serve()
	if err != nil {
		log.Fatal(err)
	}

}
