package http

import "errors"

var (
	errEnapterAPIURLEmptyOrMissing = errors.New("Enapter API URL empty or missing")
	errUnexpectedStatusCode        = errors.New("unexpected status code")
)
