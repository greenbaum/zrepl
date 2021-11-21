// Code generated by "enumer -type=PlaceholderCreationEncryptionProperty -trimprefix=PlaceholderCreationEncryptionProperty"; DO NOT EDIT.

//
package endpoint

import (
	"fmt"
)

const (
	_PlaceholderCreationEncryptionPropertyName_0 = "UnspecifiedInherit"
	_PlaceholderCreationEncryptionPropertyName_1 = "Off"
)

var (
	_PlaceholderCreationEncryptionPropertyIndex_0 = [...]uint8{0, 11, 18}
	_PlaceholderCreationEncryptionPropertyIndex_1 = [...]uint8{0, 3}
)

func (i PlaceholderCreationEncryptionProperty) String() string {
	switch {
	case 1 <= i && i <= 2:
		i -= 1
		return _PlaceholderCreationEncryptionPropertyName_0[_PlaceholderCreationEncryptionPropertyIndex_0[i]:_PlaceholderCreationEncryptionPropertyIndex_0[i+1]]
	case i == 4:
		return _PlaceholderCreationEncryptionPropertyName_1
	default:
		return fmt.Sprintf("PlaceholderCreationEncryptionProperty(%d)", i)
	}
}

var _PlaceholderCreationEncryptionPropertyValues = []PlaceholderCreationEncryptionProperty{1, 2, 4}

var _PlaceholderCreationEncryptionPropertyNameToValueMap = map[string]PlaceholderCreationEncryptionProperty{
	_PlaceholderCreationEncryptionPropertyName_0[0:11]:  1,
	_PlaceholderCreationEncryptionPropertyName_0[11:18]: 2,
	_PlaceholderCreationEncryptionPropertyName_1[0:3]:   4,
}

// PlaceholderCreationEncryptionPropertyString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func PlaceholderCreationEncryptionPropertyString(s string) (PlaceholderCreationEncryptionProperty, error) {
	if val, ok := _PlaceholderCreationEncryptionPropertyNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to PlaceholderCreationEncryptionProperty values", s)
}

// PlaceholderCreationEncryptionPropertyValues returns all values of the enum
func PlaceholderCreationEncryptionPropertyValues() []PlaceholderCreationEncryptionProperty {
	return _PlaceholderCreationEncryptionPropertyValues
}

// IsAPlaceholderCreationEncryptionProperty returns "true" if the value is listed in the enum definition. "false" otherwise
func (i PlaceholderCreationEncryptionProperty) IsAPlaceholderCreationEncryptionProperty() bool {
	for _, v := range _PlaceholderCreationEncryptionPropertyValues {
		if i == v {
			return true
		}
	}
	return false
}
