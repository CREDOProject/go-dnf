package godnf

import (
	"errors"
	"regexp"

	"github.com/CREDOProject/sharedutils/files"
)

const dnf = "dnf"

var dnfFileRegex = regexp.MustCompile(`^pip3(\.\d\d?)?\.?(\.\d\d?)?$`)

// Function used to find the dnf binary in the system.
func DetectdnfBinary() (string, error) {
	return execCommander().LookPath(dnf)
}

func DnfBinaryFrom(path string) (string, error) {
	execs, err := files.ExecsInPath(path, looksLikednf)
	if err != nil {
		return "", err
	}
	if len(execs) < 1 {
		return "", errors.New("No dnf found.")
	}

	return execs[0], err

}

// looksLikednf returns true if the given filename looks like a Dnf executable.
func looksLikednf(name string) bool {
	return dnfFileRegex.MatchString(name)
}
