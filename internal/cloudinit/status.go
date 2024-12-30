package cloudinit

import (
	"encoding/json"
	"fmt"
	"io"
)

// Status defines different possible values for the status of cloud-init in an instance.
type Status string

const (
	// StatusUnknown means that it was impossible to retrieve the status of cloud-init,
	// perhaps because the "cloud-init status" command was not successful.
	StatusUnknown Status = "Unknown"

	// StatusDone means that cloud-init completed successfully.
	StatusDone Status = "Done"

	// StatusRunning means that cloud-init is stil running on the instance.
	StatusRunning Status = "Running"

	// StatusError means that cloud-init failed.
	StatusError Status = "Error"
)

// ParseStatus checks the cloud-init status of an instance. It accepts a reader that fetches the /run/cloud-init/status.json file contents.
//
// ParseStatus returns one of the following:
//
// - (StatusDone, nil)
// - (StatusRunning, nil)
// - (StatusError, nil)
// - (StatusUnknown, <error describing why status is unknown>)
func ParseStatus(reader io.Reader) (Status, error) {
	if reader == nil {
		return StatusUnknown, fmt.Errorf("empty status.json data")
	}
	raw := &statusJSON{}
	if err := json.NewDecoder(reader).Decode(&raw); err != nil {
		return StatusUnknown, fmt.Errorf("failed to parse status.json: %w", err)
	}

	switch {
	case len(raw.V1.Init.Errors)+len(raw.V1.Init.Errors)+len(raw.V1.ModulesConfig.Errors)+len(raw.V1.ModulesFinal.Errors) > 0:
		return StatusError, nil
	case raw.V1.Stage != nil:
		return StatusRunning, nil
	}

	return StatusDone, nil
}
