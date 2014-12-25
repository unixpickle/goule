package goule

import "errors"

var (
	ErrNameTaken         = errors.New("Name already in use.")
	ErrUnknownAPI        = errors.New("Unknown API.")
	ErrRuleNotFound      = errors.New("Rule not found.")
	ErrServiceNotFound   = errors.New("Service not found.")
	ErrPermissionsDenied = errors.New("Permissions denied.")
	ErrArgumentCount     = errors.New("Invalid argument count.")
)
