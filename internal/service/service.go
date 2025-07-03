package service

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"os"
	"salesforge-assignment/internal/api"
	apierrors "salesforge-assignment/internal/api-errors"
	"salesforge-assignment/internal/config"
	"salesforge-assignment/internal/middleware/auth"
	"salesforge-assignment/internal/model"
	"salesforge-assignment/internal/repository"
)

type FormService interface {
	LoginUser(ctx context.Context, req api.Authentication) (string, error)
	CreateForm(ctx context.Context, req api.FormCreate) (*api.SelfId, error)
	GetFormById(ctx context.Context, id string) (*api.FormResponseGet, error)
	UpdateFormById(ctx context.Context, id string, update api.FormUpdate) (*api.FormResponseGet, error)
	UpdateFormStepById(ctx context.Context, formId string, stepId string, req api.FormStepUpdate) (*api.FormStepResponseGet, error)
	GetFormStepById(ctx context.Context, formId string, stepId string) (*api.FormStepResponseGet, error)
	DeleteFormStepById(ctx context.Context, formId string, stepId string) error
}

type FormServiceImpl struct {
	log                   *zerolog.Logger
	credentialsRepository repository.CredentialsRepository
	formRepository        repository.FormRepository
	config                *config.Config
	jwtKey                []byte
}

func NewFormService(
	log *zerolog.Logger,
	credentialsRepository repository.CredentialsRepository,
	formRepository repository.FormRepository,
	config *config.Config,
) FormService {
	jwtSecret := os.Getenv("JWT_SECRET_KEY")

	if jwtSecret == "" {
		log.Fatal().Msg("Environment variable JWT_SECRET_KEY is not set")
	}

	return &FormServiceImpl{
		log:                   log,
		credentialsRepository: credentialsRepository,
		formRepository:        formRepository,
		config:                config,
		jwtKey:                []byte(jwtSecret),
	}
}

func (s *FormServiceImpl) LoginUser(ctx context.Context, req api.Authentication) (string, error) {
	log := s.log.With().Str("username", req.Username).Logger()

	user, err := s.credentialsRepository.GetCredentialsByUsername(ctx, req.Username)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Debug().Msg("Password mismatch attempt")
		return "", &apierrors.InvalidCredentialsError{}
	}

	token, err := auth.GenerateToken(user.ID, user.Username, s.jwtKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT token")
		return "", &apierrors.InvalidApplicationStateError{}
	}

	log.Debug().Msg("User authenticated successfully")

	return token, nil
}

func (s *FormServiceImpl) CreateForm(ctx context.Context, req api.FormCreate) (*api.SelfId, error) {
	newForm := &model.FormModel{
		Name:                 req.Name,
		OpenTrackingEnabled:  req.OpenTrackingEnabled,
		ClickTrackingEnabled: req.ClickTrackingEnabled,
		Steps:                make([]model.FormStepModel, len(req.Steps)),
	}

	for i, step := range req.Steps {
		newForm.Steps[i] = model.FormStepModel{
			Name:      step.Name,
			Content:   step.Content,
			StepOrder: step.Step,
		}
	}

	createdForm, err := s.formRepository.CreateForm(ctx, newForm)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create form")
		return nil, &apierrors.InvalidApplicationStateError{}
	}

	log.Debug().Msg("Form created successfully")
	return &api.SelfId{
		Id:   createdForm.ID,
		Href: model.GetFormHref(createdForm.ID, s.config.Server.PublicUrl, s.config.Server.BaseURL),
	}, nil
}

func (s *FormServiceImpl) GetFormById(ctx context.Context, id string) (*api.FormResponseGet, error) {
	form, err := s.formRepository.GetFormById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug().Str("formId", id).Msg("Form not found")
			return nil, &apierrors.ResourceNotFoundError{}
		} else {
			log.Error().Err(err).Str("formId", id).Msg("Failed to retrieve form")
			return nil, &apierrors.InvalidApplicationStateError{}
		}
	}

	log.Debug().Msg("Form retrieved successfully")
	return form.ToResponse(s.config.Server.PublicUrl, s.config.Server.BaseURL), nil
}

func (s *FormServiceImpl) UpdateFormById(ctx context.Context, id string, req api.FormUpdate) (*api.FormResponseGet, error) {
	form, err := s.getFormById(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ClickTrackingEnabled == nil && req.OpenTrackingEnabled == nil {
		log.Debug().Msg("No fields to update in form")
		return nil, &apierrors.InvalidInputError{}
	}

	if req.ClickTrackingEnabled != nil {
		form.ClickTrackingEnabled = req.ClickTrackingEnabled
	}

	if req.OpenTrackingEnabled != nil {
		form.OpenTrackingEnabled = req.OpenTrackingEnabled
	}

	updatedForm, err := s.formRepository.UpdateForm(ctx, form)
	if err != nil {
		log.Error().Err(err).Str("formId", id).Msg("Failed to update form")
		return nil, &apierrors.InvalidApplicationStateError{}
	}

	return updatedForm.ToResponse(s.config.Server.PublicUrl, s.config.Server.BaseURL), nil
}

func (s *FormServiceImpl) GetFormStepById(
	ctx context.Context,
	formId string,
	stepId string,
) (*api.FormStepResponseGet, error) {
	step, err := s.getFormStepById(ctx, formId, stepId)
	if err != nil {
		return nil, err
	}

	log.Debug().Msg("Form step retrieved successfully")
	return step.ToResponse(s.config.Server.PublicUrl, s.config.Server.BaseURL), nil
}

func (s *FormServiceImpl) UpdateFormStepById(
	ctx context.Context,
	formId string,
	stepId string,
	req api.FormStepUpdate,
) (*api.FormStepResponseGet, error) {
	step, err := s.getFormStepById(ctx, formId, stepId)
	if err != nil {
		return nil, err
	}

	if req.Name == nil && req.Content == nil {
		log.Debug().Msg("No fields to update in form step")
		return nil, &apierrors.InvalidInputError{}
	}

	if req.Name != nil {
		step.Name = *req.Name
	}
	if req.Content != nil {
		step.Content = *req.Content
	}

	updatedStep, err := s.formRepository.UpdateFormStep(ctx, step)
	if err != nil {
		log.Error().Err(err).Str("stepId", stepId).Msg("Failed to update form step")
		return nil, &apierrors.InvalidApplicationStateError{}
	}

	log.Debug().Msg("Form step updated successfully")
	return updatedStep.ToResponse(s.config.Server.PublicUrl, s.config.Server.BaseURL), nil
}

func (s *FormServiceImpl) DeleteFormStepById(
	ctx context.Context,
	formId string,
	stepId string,
) error {
	step, err := s.getFormStepById(ctx, formId, stepId)
	if err != nil {
		return err
	}

	if step.FormID != formId {
		log.Debug().Str("stepId", stepId).Msg("Step does not belong to the specified form")
		return &apierrors.ResourceNotFoundError{}
	}

	err = s.formRepository.DeleteFormStepById(ctx, step.ID)
	if err != nil {
		log.Error().Err(err).Str("stepId", stepId).Msg("Failed to delete form step")
		return &apierrors.InvalidApplicationStateError{}
	}

	log.Debug().Msg("Form step deleted successfully")
	return nil
}

func (s *FormServiceImpl) getFormStepById(
	ctx context.Context,
	formId string,
	stepId string,
) (*model.FormStepModel, error) {
	step, err := s.formRepository.GetFormStepById(ctx, stepId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug().Str("stepId", stepId).Msg("Step not found")
			return nil, &apierrors.ResourceNotFoundError{}
		}
		log.Error().Err(err).Str("stepId", stepId).Msg("Failed to retrieve step")
		return nil, &apierrors.InvalidApplicationStateError{}
	}

	if step.FormID != formId {
		log.Debug().Str("stepId", stepId).Msg("Step does not belong to the specified form")
		return nil, &apierrors.ResourceNotFoundError{}
	}

	log.Debug().Msg("Step retrieved successfully")
	return step, nil
}

func (s *FormServiceImpl) getFormById(ctx context.Context, id string) (*model.FormModel, error) {
	form, err := s.formRepository.GetFormById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug().Str("formId", id).Msg("Form not found")
			return nil, &apierrors.ResourceNotFoundError{}
		}
		log.Error().Err(err).Str("formId", id).Msg("Failed to retrieve form")
		return nil, &apierrors.InvalidApplicationStateError{}
	}
	return form, nil
}
