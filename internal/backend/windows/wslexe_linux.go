package windows

// This file mocks utilities to access functionality accessed via wsl.exe

import (
	"errors"

	"github.com/ubuntu/gowsl/internal/state"
)

// Shutdown shuts down all distros
// This implementation will always fail on Linux.
func (Backend) Shutdown() error {
	return errors.New("not implemented")
}

// Terminate shuts down a particular distro
// This implementation will always fail on Linux.
func (Backend) Terminate(distroName string) error {
	return errors.New("not implemented")
}

// SetAsDefault sets a particular distribution as the default one.
// This implementation will always fail on Linux.
func (Backend) SetAsDefault(distroName string) error {
	return errors.New("not implemented")
}

// State returns the state of a particular distro as seen in `wsl.exe -l -v`.
// This implementation will always fail on Linux.
func (Backend) State(distributionName string) (s state.State, err error) {
	return s, errors.New("not implemented")
}
