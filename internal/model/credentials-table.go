package model

type CredentialsModel struct {
	ID       string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username string `gorm:"type:text;not null;unique"`
	Password string `gorm:"not null;type:text"`
}

func (*CredentialsModel) TableName() string {
	return "authz.credentials"
}
