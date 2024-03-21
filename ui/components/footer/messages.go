package footer

import "strings"

// SubmitMsg is a string such as "w foo.txt" containing an action
// and arguments
type SubmitMsg string

// CancelMsg cancels the current action.
type CancelMsg struct{}

// StatusMsg contains a non-error status
type StatusMsg string

// ErrorMsg contains an error status
type ErrorMsg string

// Decode the text, returning the action and its arguments (if any)
func (m SubmitMsg) Decode() (string, []string) {
	slices := strings.Fields(string(m))
	switch len(slices) {
	case 0:
		return "", nil
	case 1:
		return slices[0], nil
	default:
		return slices[0], slices[1:]
	}
}
