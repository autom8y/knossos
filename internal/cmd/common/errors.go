package common

import (
	"github.com/autom8y/knossos/internal/errors"
	"github.com/autom8y/knossos/internal/output"
)

// PrintAndReturn prints the error using the printer and returns it wrapped
// as Handled so that main.go won't print it again.
func PrintAndReturn(printer *output.Printer, err error) error {
	printer.PrintError(err)
	return errors.Handled(err)
}
