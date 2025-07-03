package model

import (
	"fmt"
	"salesforge-assignment/internal/api"
)

const (
	formHref     = "%s%s/form/%s"
	formStepHref = "%s%s/form/%s/steps/%s"
)

type FormModel struct {
	ID                   string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OpenTrackingEnabled  *bool           `gorm:"not null;default:false"`
	ClickTrackingEnabled *bool           `gorm:"not null;default:false"`
	Name                 string          `gorm:"not null;unique"`
	Steps                []FormStepModel `gorm:"foreignKey:FormID"`
}

func (*FormModel) TableName() string {
	return "public.form"
}

func (s *FormModel) ToResponse(publicUrl string, baseUrl string) *api.FormResponseGet {
	steps := make([]api.FormStepResponseGet, 0, len(s.Steps))
	for _, step := range s.Steps {
		steps = append(steps, *step.ToResponse(publicUrl, baseUrl))
	}

	response := &api.FormResponseGet{
		Name:                 s.Name,
		OpenTrackingEnabled:  *s.OpenTrackingEnabled,
		ClickTrackingEnabled: *s.ClickTrackingEnabled,
		Steps:                steps,
		Self: api.SelfId{
			Id:   s.ID,
			Href: GetFormHref(s.ID, publicUrl, baseUrl),
		},
	}

	return response
}

func GetFormHref(formId string, publicUrl string, baseUrl string) string {
	return fmt.Sprintf(formHref, publicUrl, baseUrl, formId)
}

func GetFormStepHref(formId string, stepId string, publicUrl string, baseUrl string) string {
	return fmt.Sprintf(formStepHref, publicUrl, baseUrl, formId, stepId)
}
