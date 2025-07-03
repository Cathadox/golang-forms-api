package model

import "salesforge-assignment/internal/api"

type FormStepModel struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string `gorm:"not null;unique"`
	Content   string `gorm:"not null"`
	StepOrder int    `gorm:"not null"`
	FormID    string `gorm:"not null;type:uuid;"`
}

func (*FormStepModel) TableName() string {
	return "public.form_steps"
}

func (s *FormStepModel) ToResponse(publicUrl string, baseUrl string) *api.FormStepResponseGet {
	return &api.FormStepResponseGet{
		Name:    s.Name,
		Content: s.Content,
		Step:    s.StepOrder,
		Self: api.SelfId{
			Id:   s.ID,
			Href: GetFormStepHref(s.FormID, s.ID, publicUrl, baseUrl),
		},
	}
}
