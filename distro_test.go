package wsl_test

import (
	"context"
	"fmt"
	"testing"
	"wsl"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDistroString(t *testing.T) {
	realDistro := newTestDistro(t, jammyRootFs)
	fakeDistro := wsl.Distro{Name: sanitizeDistroName(fmt.Sprintf("%s_%s_%s", namePrefix, t.Name(), uniqueId()))}

	testCases := map[string]struct {
		fakeDistro bool
		withoutEnv bool
		wants      string
	}{
		"nominal": {wants: fmt.Sprintf(`distro: %s
configuration:
  - Version: 2
  - DefaultUID: 0
  - InteropEnabled: true
  - PathAppended: true
  - DriveMountingEnabled: true
  - undocumentedWSLVersion: 2
  - DefaultEnvironmentVariables:
    - HOSTTYPE: x86_64
    - LANG: en_US.UTF-8
    - PATH: /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games
    - TERM: xterm-256color
`, realDistro.Name)},
		"wrong distro": {fakeDistro: true, wants: fmt.Sprintf(`distro: %s
configuration: failed to get configuration, failed syscall to WslGetDistributionConfiguration
`, fakeDistro.Name)},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			d := realDistro
			if tc.fakeDistro {
				d = fakeDistro
			}

			got := d.String()
			require.Equal(t, tc.wants, got)
		})
	}
}

// One table test per setting, with:
//Set a value to non default, assert with Configuration()
// Set a value fails, assert the configuration didn't change:
// -- error no distro
// -- error on null character name in distro

// TestConfigurationWithChanges
// set multiple non default values, assert with Configuration()

// TestConfiguration:
//Assert default
// Error cases:
// - no distro
// - null character name in distro
func TestConfiguration(t *testing.T) {

	distro := newTestDistro(t, jammyRootFs)

	cmd := distro.Command(context.Background(), "useradd testuser")
	err := cmd.Run()
	require.NoError(t, err)

	default_config, err := distro.GetConfiguration()
	require.NoError(t, err)

	tests := map[string]struct {
		defaultUID           uint32
		interopEnabled       bool
		pathAppended         bool
		driveMountingEnabled bool
	}{
		// Root user cases
		"Root":                                                  {defaultUID: 0},
		"Root DriveMountingEnabled":                             {defaultUID: 0, driveMountingEnabled: true},
		"Root PathAppended":                                     {defaultUID: 0, pathAppended: true},
		"Root PathAppended DriveMountingEnabled":                {defaultUID: 0, pathAppended: true, driveMountingEnabled: true},
		"Root InteropEnabled":                                   {defaultUID: 0, interopEnabled: true},
		"Root InteropEnabled DriveMountingEnabled":              {defaultUID: 0, interopEnabled: true, driveMountingEnabled: true},
		"Root InteropEnabled PathAppended":                      {defaultUID: 0, interopEnabled: true, pathAppended: true},
		"Root InteropEnabled PathAppended DriveMountingEnabled": {defaultUID: 0, interopEnabled: true, pathAppended: true, driveMountingEnabled: true},

		// Default user cases
		"User":                                                  {defaultUID: 1000},
		"User DriveMountingEnabled":                             {defaultUID: 1000, driveMountingEnabled: true},
		"User PathAppended":                                     {defaultUID: 1000, pathAppended: true},
		"User PathAppended DriveMountingEnabled":                {defaultUID: 1000, pathAppended: true, driveMountingEnabled: true},
		"User InteropEnabled":                                   {defaultUID: 1000, interopEnabled: true},
		"User InteropEnabled DriveMountingEnabled":              {defaultUID: 1000, interopEnabled: true, driveMountingEnabled: true},
		"User InteropEnabled PathAppended":                      {defaultUID: 1000, interopEnabled: true, pathAppended: true},
		"User InteropEnabled PathAppended DriveMountingEnabled": {defaultUID: 1000, interopEnabled: true, pathAppended: true, driveMountingEnabled: true},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			defer func() { // Resetting to default state
				distro.DefaultUID(default_config.DefaultUID)
				distro.InteropEnabled(default_config.InteropEnabled)
				distro.PathAppended(default_config.PathAppended)
				distro.DriveMountingEnabled(default_config.DriveMountingEnabled)
			}()

			err = distro.DefaultUID(tc.defaultUID)
			require.NoError(t, err)

			err = distro.InteropEnabled(tc.interopEnabled)
			require.NoError(t, err)

			err = distro.PathAppended(tc.pathAppended)
			require.NoError(t, err)

			err = distro.DriveMountingEnabled(tc.driveMountingEnabled)
			require.NoError(t, err)

			got, err := distro.GetConfiguration()
			require.NoError(t, err)

			// Config test
			assert.Equal(t, tc.defaultUID, got.DefaultUID)
			assert.Equal(t, tc.interopEnabled, got.InteropEnabled)
			assert.Equal(t, tc.pathAppended, got.PathAppended)
			assert.Equal(t, tc.driveMountingEnabled, got.DriveMountingEnabled)
		})
	}
}
func TestGetConfiguration(t *testing.T) {
	d := newTestDistro(t, jammyRootFs)
	c, err := d.GetConfiguration()
	require.NoError(t, err)
	assert.Equal(t, c.Version, uint8(2))
	assert.Equal(t, c.DefaultUID, uint32(0))
	assert.Equal(t, c.InteropEnabled, true)
	assert.Equal(t, c.PathAppended, true)
	assert.Equal(t, c.DriveMountingEnabled, true)

	defaultEnvs := map[string]string{
		"HOSTTYPE": "x86_64",
		"LANG":     "en_US.UTF-8",
		"PATH":     "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games",
		"TERM":     "xterm-256color",
	}
	assert.Equal(t, c.DefaultEnvironmentVariables, defaultEnvs)
}

func TestGetConfigurationDistroError(t *testing.T) {
	d := wsl.Distro{Name: "I'm not registered"}
	_, err := d.GetConfiguration()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed syscall")
}

func TestGetConfigurationNameError(t *testing.T) {
	d := wsl.Distro{Name: "I'm not a \x00 valid string"}
	_, err := d.GetConfiguration()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to convert")
}
