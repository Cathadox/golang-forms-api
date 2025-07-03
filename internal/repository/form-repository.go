package repository

import (
	"context"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"salesforge-assignment/internal/model"
)

type FormRepository interface {
	CreateForm(ctx context.Context, form *model.FormModel) (*model.FormModel, error)
	GetFormById(ctx context.Context, id string) (*model.FormModel, error)
	UpdateForm(ctx context.Context, form *model.FormModel) (*model.FormModel, error)
	UpdateFormStep(ctx context.Context, step *model.FormStepModel) (*model.FormStepModel, error)
	DeleteFormStepById(ctx context.Context, id string) error
	GetFormStepById(ctx context.Context, formId string) (*model.FormStepModel, error)
	DeleteFormStep(ctx context.Context, step *model.FormStepModel) error
}

type FormRepositoryImpl struct {
	log *zerolog.Logger
	db  *gorm.DB
}

func NewFormRepository(
	log *zerolog.Logger,
	db *gorm.DB,
) FormRepository {
	return &FormRepositoryImpl{
		log: log,
		db:  db,
	}
}

func (sr *FormRepositoryImpl) CreateForm(ctx context.Context, form *model.FormModel) (*model.FormModel, error) {
	err := sr.db.WithContext(ctx).Create(&form).Error
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (sr *FormRepositoryImpl) GetFormById(ctx context.Context, id string) (*model.FormModel, error) {
	var form *model.FormModel
	err := sr.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order ASC")
		}).
		First(&form, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (sr *FormRepositoryImpl) UpdateForm(ctx context.Context, form *model.FormModel) (*model.FormModel, error) {
	err := sr.db.WithContext(ctx).Save(form).Error
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (sr *FormRepositoryImpl) UpdateFormStep(ctx context.Context, step *model.FormStepModel) (*model.FormStepModel, error) {
	err := sr.db.WithContext(ctx).Save(step).Error
	if err != nil {
		return nil, err
	}
	return step, nil
}

func (sr *FormRepositoryImpl) DeleteFormStepById(ctx context.Context, id string) error {
	err := sr.db.WithContext(ctx).Where("id = ?", id).Delete(&model.FormStepModel{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (sr *FormRepositoryImpl) GetFormStepById(ctx context.Context, stepId string) (*model.FormStepModel, error) {
	var step *model.FormStepModel
	err := sr.db.WithContext(ctx).Where("id = ?", stepId).Find(&step).Error
	if err != nil {
		return nil, err
	}
	return step, nil
}

func (sr *FormRepositoryImpl) DeleteFormStep(ctx context.Context, step *model.FormStepModel) error {
	err := sr.db.WithContext(ctx).Delete(step).Error
	if err != nil {
		sr.log.Error().Err(err).Msg("Failed to delete form step")
		return err
	}
	return nil
}
