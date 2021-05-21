// +build !windows

package standalone

import (
	"fmt"
)

// Stop terminates the application process.
func Stop(appID string) error {
	apps, err := List()
	if err != nil {
		return err
	}

	for _, a := range apps {
		if a.AppID == appID {
			pid := fmt.Sprintf("%v", a.PID)

			_, err := RunCmdAndWait("kill", pid)

			return err
		}
	}

	return fmt.Errorf("couldn't find app id %s", appID)
}
