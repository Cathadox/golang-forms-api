package unit

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"salesforge-assignment/internal/api"
	apierrors "salesforge-assignment/internal/api-errors"
	"salesforge-assignment/internal/handler"
	"salesforge-assignment/internal/middleware"
	"testing"
)

type fakeFieldError struct {
	field string
	tag   string
}

func (f fakeFieldError) Error() string                    { return "" }
func (f fakeFieldError) Field() string                    { return f.field }
func (f fakeFieldError) Tag() string                      { return f.tag }
func (f fakeFieldError) ActualTag() string                { return f.tag }
func (f fakeFieldError) Namespace() string                { return f.field }
func (f fakeFieldError) StructNamespace() string          { return f.field }
func (f fakeFieldError) StructField() string              { return f.field }
func (f fakeFieldError) Value() interface{}               { return nil }
func (f fakeFieldError) Param() string                    { return "" }
func (f fakeFieldError) Kind() reflect.Kind               { return reflect.String }
func (f fakeFieldError) Type() reflect.Type               { return reflect.TypeOf("") }
func (f fakeFieldError) Translate(_ ut.Translator) string { return "" }

func setupContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// directly set the logger in context
	l := zerolog.Nop()
	c.Set(middleware.LoggerKey, &l)
	return c, w
}

func TestHandleError_ValidationError(t *testing.T) {
	errs := validator.ValidationErrors{fakeFieldError{"Name", "required"}}
	c, w := setupContext()

	handler.HandleError(c, errs)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body api.ValidationErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, 400, body.Code)
	assert.Equal(t, "One or more fields failed validation.", body.Message)
	assert.Len(t, body.Errors, 1)
	assert.Equal(t, "Name", body.Errors[0].Field)
	assert.Contains(t, body.Errors[0].Message, "required")
}

func TestHandleError_HTTPErrorTypes(t *testing.T) {
	types := []struct {
		err  apierrors.HTTPError
		code int
		msg  string
	}{
		{&apierrors.InvalidApplicationStateError{}, 400, "Invalid application state"},
		{&apierrors.PermissionDeniedError{}, 403, "Permission denied"},
		{&apierrors.InvalidInputError{}, 400, "Invalid input"},
		{&apierrors.ResourceNotFoundError{}, 404, "Resource not found"},
		{&apierrors.InvalidCredentialsError{}, 401, "Invalid credentials"},
		{&apierrors.UnauthorizedError{}, 401, "Unauthorized access"},
	}

	for _, tt := range types {
		c, w := setupContext()

		handler.HandleError(c, tt.err)

		resp := w.Result()
		assert.Equal(t, tt.code, resp.StatusCode)

		var body api.ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, tt.msg, body.Message)
	}
}

func TestHandleError_InvalidRequestBodyError(t *testing.T) {
	c1, w1 := setupContext()
	err1 := &apierrors.InvalidRequestBodyError{Err: nil}
	handler.HandleError(c1, err1)

	resp1 := w1.Result()
	assert.Equal(t, http.StatusBadRequest, resp1.StatusCode)

	var body1 api.ErrorResponse
	err := json.NewDecoder(resp1.Body).Decode(&body1)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request body.", body1.Message)

	orig := errors.New("oops")
	c2, w2 := setupContext()
	err2 := &apierrors.InvalidRequestBodyError{Err: orig}
	handler.HandleError(c2, err2)

	resp2 := w2.Result()
	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

	var body2 api.ErrorResponse
	err = json.NewDecoder(resp2.Body).Decode(&body2)
	assert.NoError(t, err)
	assert.Contains(t, body2.Message, "Invalid request body.")
}

func TestHandleError_DefaultError(t *testing.T) {
	c, w := setupContext()
	handler.HandleError(c, errors.New("unknown"))

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body api.ErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid application state", body.Message)
}
