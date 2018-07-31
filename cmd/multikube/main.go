package main

import (
	"os"
	"fmt"
	"log"
	"github.com/spf13/pflag"
	"gitlab.com/amimof/multikube"
)

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

	// Create the proxy
	p := multikube.NewProxy()

	// Create the server
	s := multikube.NewServer(p)

	err := s.Serve()
	if err != nil {
		log.Fatal(err)
	}

}
