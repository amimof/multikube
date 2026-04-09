package main

import (
	"context"
	"os"
	"time"

	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

func newEditCertificateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certificate [NAME]",
		Short: "Edit a certificate",
		Long:  `Edit a certificate`,
		Example: `
  # Edit a certificate
  multikubectl edit certificate prod-certificate-v1`,
		Args: cobra.ExactArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			return runEditCertificateCmd(cmd, args)
		}),
	}

	return cmd
}

func runEditCertificateCmd(cmd *cobra.Command, args []string) error {
	baseCtx := cmd.Context()

	tracer := otel.Tracer("multikubectl")
	baseCtx, span := tracer.Start(baseCtx, "multikubectl.edit.certificate")
	defer span.End()

	nname := args[0]

	getCtx, certificatencel := context.WithTimeout(baseCtx, time.Second*30)
	defer certificatencel()

	certificate, err := clientSet.CertificateV1().Get(getCtx, nname)
	if err != nil {
		return err
	}

	var updatedCertificate certificatev1.Certificate
	err = runEditor(certificate, &updatedCertificate)
	if err != nil {
		return err
	}

	// Exit early if no changes where made
	if proto.Equal(certificate, &updatedCertificate) {
		logrus.Info("no changes detected")
		os.Exit(0)
	}

	// Send update to server
	updateCtx, certificatencel := context.WithTimeout(baseCtx, time.Second*30)
	defer certificatencel()
	err = clientSet.CertificateV1().Update(updateCtx, nname, &updatedCertificate)
	if err != nil {
		return err
	}

	logrus.Infof("certificate %s was updated", nname)

	return nil
}
