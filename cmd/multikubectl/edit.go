package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/amimof/multikube/pkg/cmdutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func newEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit resources",
		Long:  `Edit Get resources`,
		Example: `
# Edit a backend
multikube edit backend test-backend

# Edit a backend in json format
multikube edit backend test-backend -o json
`,
		Args: cobra.ExactArgs(1),
	}

	cmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format")

	cmd.AddCommand(newEditBackendCmd())
	cmd.AddCommand(newEditCACmd())
	cmd.AddCommand(newEditCertificateCmd())
	cmd.AddCommand(newEditCredentialCmd())
	cmd.AddCommand(newEditPolicyCmd())
	cmd.AddCommand(newEditRouteCmd())

	return cmd
}

func runEditor(m, dst proto.Message) error {
	codec, err := cmdutil.CodecFor(outputFormat)
	if err != nil {
		logrus.Fatalf("error creating serializer: %v", err)
	}

	b, err := codec.Serialize(m)
	if err != nil {
		logrus.Fatalf("error serializing: %v", err)
	}

	// Create temporary file to hold the JSON
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("*.%s", outputFormat))
	if err != nil {
		return err
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			logrus.Fatalf("error closing tmp file: %v", err)
		}
	}()

	_, err = tmpFile.Write(b)
	if err != nil {
		return err
	}

	// Get the editor from the environment variable, default to Vim
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// Open text editor
	editorCmd := exec.Command(editor, tmpFile.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return err
	}

	// Read modified ctr in file
	ub, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return err
	}

	return codec.Deserialize(ub, dst)
}
