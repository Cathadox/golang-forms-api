package unit

import (
	"github.com/stretchr/testify/assert"
	"salesforge-assignment/internal/model"
	"testing"
)

func TestGetFormHref(t *testing.T) {
	seqID := "seq123"
	publicURL := "https://public.example"
	baseURL := "/api/v1"
	expected := "https://public.example/api/v1/form/seq123"
	href := model.GetFormHref(seqID, publicURL, baseURL)
	assert.Equal(t, expected, href)
}

func TestGetFormStepHref(t *testing.T) {
	seqID := "seq123"
	stepID := "step456"
	publicURL := "https://public.example"
	baseURL := "/api/v1"
	expected := "https://public.example/api/v1/form/seq123/steps/step456"
	href := model.GetFormStepHref(seqID, stepID, publicURL, baseURL)
	assert.Equal(t, expected, href)
}

func TestFormModel_ToResponse_NoSteps(t *testing.T) {
	openTrack := true
	clickTrack := false
	seq := &model.FormModel{
		ID:                   "id1",
		Name:                 "Test Form",
		OpenTrackingEnabled:  &openTrack,
		ClickTrackingEnabled: &clickTrack,
		Steps:                []model.FormStepModel{},
	}

	publicURL := "https://public.example"
	baseURL := "/api/v1"

	resp := seq.ToResponse(publicURL, baseURL)
	assert := assert.New(t)

	assert.Equal(seq.Name, resp.Name)
	assert.Equal(openTrack, resp.OpenTrackingEnabled)
	assert.Equal(clickTrack, resp.ClickTrackingEnabled)
	assert.Equal(seq.ID, resp.Self.Id)
	expectedHref := model.GetFormHref(seq.ID, publicURL, baseURL)
	assert.Equal(expectedHref, resp.Self.Href)
	// No steps
	assert.Empty(resp.Steps)
}

func TestFormModel_ToResponse_WithSteps(t *testing.T) {
	openTrack := false
	clickTrack := true
	step1 := model.FormStepModel{FormID: "id1", ID: "s1"}
	step2 := model.FormStepModel{FormID: "id1", ID: "s2"}
	seq := &model.FormModel{
		ID:                   "id1",
		Name:                 "Test Form",
		OpenTrackingEnabled:  &openTrack,
		ClickTrackingEnabled: &clickTrack,
		Steps:                []model.FormStepModel{step1, step2},
	}

	publicURL := "https://public.example"
	baseURL := "/api/v1"
	resp := seq.ToResponse(publicURL, baseURL)

	assert := assert.New(t)
	assert.Len(resp.Steps, 2)
	// Verify each step's Self href uses the correct form and step IDs
	for i, stepResp := range resp.Steps {
		expectedStepHref := model.GetFormStepHref(seq.ID, stepResp.Self.Id, publicURL, baseURL)
		assert.Equal(expectedStepHref, stepResp.Self.Href)
		assert.Equal(stepResp.Self, resp.Steps[i].Self)
	}
}

func TestFormStepModel_ToResponse(t *testing.T) {
	step := &model.FormStepModel{
		FormID:    "seqX",
		ID:        "stepX",
		Name:      "Standalone",
		Content:   "Content here",
		StepOrder: 5,
	}

	publicURL := "https://public.example"
	baseURL := "/api/v1"
	resp := step.ToResponse(publicURL, baseURL)
	assert := assert.New(t)

	assert.Equal(step.Name, resp.Name)
	assert.Equal(step.Content, resp.Content)
	assert.Equal(step.StepOrder, resp.Step)
	assert.Equal(step.ID, resp.Self.Id)
	expectedHref := model.GetFormStepHref(step.FormID, step.ID, publicURL, baseURL)
	assert.Equal(expectedHref, resp.Self.Href)
}
