package api_errors

import (
	"errors"
	"fmt"
	"salesforge-assignment/internal/api"
)

type HTTPError interface {
	error
	APIErrorResponse() api.ErrorResponse
}

type InvalidApplicationStateError struct {
	Err error
}

func (err *InvalidApplicationStateError) Error() string {
	return "invalid application state"
}

func (err *InvalidApplicationStateError) APIErrorResponse() api.ErrorResponse {
	return api.ErrorResponse{
		Message: "Invalid application state",
		Code:    500,
	}
}

type PermissionDeniedError struct {
	Err error
}

func (err *PermissionDeniedError) Error() string {
	return "permission denied"
}

func (err *PermissionDeniedError) APIErrorResponse() api.ErrorResponse {
	return api.ErrorResponse{
		Message: "Permission denied",
		Code:    403,
	}
}

type InvalidInputError struct {
	Err error
}

func (err *InvalidInputError) Error() string {
	return "invalid input"
}

func (err *InvalidInputError) APIErrorResponse() api.ErrorResponse {
	return api.ErrorResponse{
		Message: "Invalid input",
		Code:    400,
	}
}

type ResourceNotFoundError struct {
	Err error
}

func (err *ResourceNotFoundError) Error() string {
	return "resource not found"
}

func (err *ResourceNotFoundError) APIErrorResponse() api.ErrorResponse {
	return api.ErrorResponse{
		Message: "Resource not found",
		Code:    404,
	}
}

type InvalidCredentialsError struct {
	Err error
}

func (err *InvalidCredentialsError) Error() string {
	return "invalid credentials"
}

func (err *InvalidCredentialsError) APIErrorResponse() api.ErrorResponse {
	return api.ErrorResponse{
		Message: "Invalid credentials",
		Code:    401,
	}
}

type UnauthorizedError struct {
	Err error
}

func (err *UnauthorizedError) Error() string {
	return "Unauthorized access"
}

func (err *UnauthorizedError) APIErrorResponse() api.ErrorResponse {
	return api.ErrorResponse{
		Message: "Unauthorized access",
		Code:    401,
	}
}

type InvalidRequestBodyError struct {
	Err error
}

func (err *InvalidRequestBodyError) Error() string {
	return "invalid request body"
}

func (err *InvalidRequestBodyError) APIErrorResponse() api.ErrorResponse {
	originalErr := errors.Unwrap(err.Err)
	var message string
	if originalErr != nil {
		message = fmt.Sprintf("Invalid request body: %s", originalErr.Error())
	} else {
		message = "Invalid request body."
	}

	return api.ErrorResponse{
		Message: message,
		Code:    400,
	}
}

func (err *InvalidRequestBodyError) Unwrap() error {
	return err.Err
}
