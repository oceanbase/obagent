package shellf

import "regexp"

var (
	packageNameRegex = regexp.MustCompile(`^([a-zA-Z0-9\-]+)(-(\d+.\d+.\d+(.\d+)?)(-([\d.]+)\.([a-zA-Z0-9]+)\.([a-zA-Z0-9_]+))?)?$`)
)

func (t CommandParameterType) Validate(value string) bool {
	switch t {
	case packageNameType:
		return packageNameRegex.MatchString(value)
	default:
		return true
	}
}
