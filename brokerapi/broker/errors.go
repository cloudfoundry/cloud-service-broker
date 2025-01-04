package broker

import (
	"errors"
	"net/http"

	"github.com/pivotal-cf/brokerapi/v11/domain/apiresponses"
)

// typed errors which are not declared in pivotal-cf/brokerapi

var (
	badRequestMsg             = "bad request"
	invalidUserInputMsg       = "User supplied parameters must be in the form of a valid JSON map."
	nonUpdateableParameterMsg = "attempt to update parameter that may result in service instance re-creation and data loss"
	notFoundMsg               = "not found"
	concurrencyErrorMsg       = "ConcurrencyError"

	badRequestKey            = "bad-request"
	invalidUserInputKey      = "parsing-user-request"
	nonUpdatableParameterKey = "prohibited"
	notFoundKey              = "not-found"
	concurrencyErrorKey      = "concurrency-error"

	ErrBadRequest            = apiresponses.NewFailureResponse(errors.New(badRequestMsg), http.StatusBadRequest, badRequestKey)
	ErrInvalidUserInput      = apiresponses.NewFailureResponse(errors.New(invalidUserInputMsg), http.StatusBadRequest, invalidUserInputKey)
	ErrNonUpdatableParameter = apiresponses.NewFailureResponse(errors.New(nonUpdateableParameterMsg), http.StatusBadRequest, nonUpdatableParameterKey)
	ErrNotFound              = apiresponses.NewFailureResponse(errors.New(notFoundMsg), http.StatusNotFound, notFoundKey)
	ErrConcurrencyError      = apiresponses.NewFailureResponse(errors.New(concurrencyErrorMsg), http.StatusUnprocessableEntity, concurrencyErrorKey)
)
