// This package contains some helper functions of the various internal packages.
package utils

import "strconv"

// Returns a default value if the initial string is empty string.
//
// (Ie. initial == "" returns the default value.)
func DefaultStringValue(initial string, defaultValue string) string {
	if initial != "" {
		return initial
	}
	return defaultValue
}

// Takes in a string value to parse, and a pointer to an error.
//
// An int result is returned, and if there is an error,
// the error pointer is updated with the error.
//
// WARNING: Use this function carefully, the error passed
// as a pointer needs to be resolved, by the caller
// in the future.
func InlineIntParse(value string, errorPtr *error) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		*errorPtr = err
	}

	return result
}

// Takes in a string value to parse, and a pointer to an error.
//
// A bool result is returned, and if there is an error,
// the error pointer is updated with the error.
//
// WARNING: Use this function carefully, the error passed
// as a pointer needs to be resolved, by the caller
// in the future.
func InlineBoolParse(value string, errorPtr *error) bool {
	result, err := strconv.ParseBool(value)
	if err != nil {
		*errorPtr = err
	}

	return result
}
