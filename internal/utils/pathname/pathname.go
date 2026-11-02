package pathname

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"
)

const folderMaxName = 255

const forbiddenChars = `<>:"/\|?*`

var (
	ErrMaxLengthExceeded = fmt.Errorf("length exceeds maximum of %d characters", folderMaxName)
	ErrEmptyName         = errors.New("can't be empty")
	ErrDotName           = errors.New(`can't be "." or ".."`)
	ErrForbiddenChar     = errors.New("can't contain any of " + forbiddenChars)
	ErrControlChar       = errors.New("can't contain control characters")
	ErrTrailingDotSpace  = errors.New("can't end with a dot or a space")
	ErrReservedName      = errors.New("is a reserved name on Windows")
)

var windowsReservedNames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

func Validate(name string) error {
	if len(name) > folderMaxName {
		return ErrMaxLengthExceeded
	}
	switch {
	case name == "":
		return ErrEmptyName
	case name == "." || name == "..":
		return ErrDotName
	case strings.ContainsAny(name, forbiddenChars):
		return ErrForbiddenChar
	case strings.IndexFunc(name, unicode.IsControl) >= 0:
		return ErrControlChar
	case strings.HasSuffix(name, ".") || strings.HasSuffix(name, " "):
		return ErrTrailingDotSpace
	case isWindowsReserved(name):
		return ErrReservedName
	}
	return nil
}

func isWindowsReserved(name string) bool {
	stem, _, _ := strings.Cut(name, ".")
	return slices.Contains(windowsReservedNames, strings.ToUpper(strings.TrimRight(stem, " ")))
}
