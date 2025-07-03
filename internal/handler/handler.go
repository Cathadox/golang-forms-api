package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"salesforge-assignment/internal/api"
	apierrors "salesforge-assignment/internal/api-errors"
	"salesforge-assignment/internal/service"
)

var _ api.ServerInterface = (*FormHandler)(nil)

type FormHandler struct {
	svc service.FormService
}

func NewFormHandler(
	svc service.FormService,
) *FormHandler {
	return &FormHandler{
		svc: svc,
	}
}

var validate = validator.New()

func (h *FormHandler) LoginUser(c *gin.Context) {
	var req api.Authentication

	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, &apierrors.InvalidRequestBodyError{Err: err})
		return
	}

	token, err := h.svc.LoginUser(c.Request.Context(), req)
	if err != nil {
		HandleError(c, err)
		return
	}

	response := api.AuthenticationResponse{
		Token: token,
	}

	c.JSON(http.StatusOK, response)
}

func (h *FormHandler) CreateForm(c *gin.Context) {

	var req api.FormCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, &apierrors.InvalidRequestBodyError{Err: err})
		return
	}

	if err := validate.Struct(&req); err != nil {
		HandleError(c, err)
		return
	}

	self, err := h.svc.CreateForm(c.Request.Context(), req)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, self)
}

func (h *FormHandler) GetFormById(c *gin.Context, formId string) {
	form, err := h.svc.GetFormById(c.Request.Context(), formId)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, form)
}

func (h *FormHandler) UpdateFormById(c *gin.Context, formId string) {
	var req api.FormUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, &apierrors.InvalidRequestBodyError{Err: err})
		return
	}

	if err := validate.Struct(&req); err != nil {
		HandleError(c, err)
		return
	}

	form, err := h.svc.UpdateFormById(c.Request.Context(), formId, req)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, form)
}

func (h *FormHandler) GetFormStepById(c *gin.Context, formId string, stepId string) {
	step, err := h.svc.GetFormStepById(c.Request.Context(), formId, stepId)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, step)
}

func (h *FormHandler) UpdateFormStepById(c *gin.Context, formId string, stepId string) {
	var req api.FormStepUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, &apierrors.InvalidRequestBodyError{Err: err})
		return
	}

	if err := validate.Struct(&req); err != nil {
		HandleError(c, err)
		return
	}

	step, err := h.svc.UpdateFormStepById(c.Request.Context(), formId, stepId, req)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, step)
}

func (h *FormHandler) DeleteFormStepById(c *gin.Context, formId string, stepId string) {
	err := h.svc.DeleteFormStepById(c.Request.Context(), formId, stepId)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
