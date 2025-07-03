package repository

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	apierrors "salesforge-assignment/internal/api-errors"
	"salesforge-assignment/internal/model"
)

type CredentialsRepository interface {
	GetCredentialsByUsername(ctx context.Context, username string) (*model.CredentialsModel, error)
}

type CredentialsRepositoryImpl struct {
	log *zerolog.Logger
	db  *gorm.DB
}

func NewCredentialsRepository(
	log *zerolog.Logger,
	db *gorm.DB,
) CredentialsRepository {
	return &CredentialsRepositoryImpl{
		log: log,
		db:  db,
	}
}

func (cr *CredentialsRepositoryImpl) GetCredentialsByUsername(ctx context.Context, username string) (*model.CredentialsModel, error) {
	var credentials model.CredentialsModel
	err := cr.db.WithContext(ctx).
		Where("username = ?", username).
		Find(&credentials).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apierrors.ResourceNotFoundError{}
		}
		return nil, &apierrors.InvalidApplicationStateError{}
	}

	cr.log.Trace().Msg("Retrieved credentials for user: " + username)
	return &credentials, nil
}
