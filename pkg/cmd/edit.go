package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var editCommand = &cobra.Command{
	Use:   "edit",
	Short: "Edit a kube secret file",
	Run:   runEditCommand,
	Args:  cobra.ExactArgs(1), // [filename.yml]
}

// This isn't particularly secure in how we use this (race conditions and what
// not), but it's good enough. Caller is responsible for `defer os.Remove(name)`
func getTempFileName(dir, pattern string) (string, error) {
	fh, err := ioutil.TempFile("", "ksed*.yml")
	if err != nil {
		return "", err
	}

	name := fh.Name()
	fh.Close()
	return name, nil
}

func runEditor(ed, f string) error {
	cmd := exec.Command(ed, f)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error invoking %q: %s", ed, err)
	}
	return err
}

func runEditCommand(cmd *cobra.Command, args []string) {
	inF := args[0]
	editor := whichEditor()

	tmpF, err := getTempFileName("", "ksed*.yml")
	if err != nil {
		errorExit(err)
	}
	defer os.Remove(tmpF)

	err = secretReadMungeWrite(inF, tmpF, "decode")
	if err != nil {
		errorExit(err)
	}

	err = runEditor(editor, tmpF)
	if err != nil {
		errorExit(err)
	}

	err = secretReadMungeWrite(inF, tmpF, "encode")
	if err != nil {
		errorExit(err)
	}

	fmt.Fprintln(os.Stderr, "All done!")
}

func init() {
	rootCmd.AddCommand(editCommand)
}
