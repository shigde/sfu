package handler

import "errors"

var (
	errAccountNameNotFound = errors.New("no account name in request found")
	errAccountIriNotFound  = errors.New("no 'accountIri' get parameter in request")
	errAccountIriInvalid   = errors.New("invalid 'accountIri'")
	errNoFederationSupport = errors.New("no federation support")
)
