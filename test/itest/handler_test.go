package itest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"net/http/httptest"
	"salesforge-assignment/internal/api"
	"salesforge-assignment/internal/model"
)

func (suite *HandlerIntegrationSuite) performRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)

		suite.Require().NoError(err, "Test setup failed: could not marshal request body to JSON")

		bodyReader = bytes.NewBuffer(jsonBody)
	} else {
		bodyReader = nil
	}

	req, err := http.NewRequest(method, "/api/v1"+path, bodyReader)
	suite.Require().NoError(err, "Test setup failed: could not create HTTP request")

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

func (suite *HandlerIntegrationSuite) getAuthTokenForTestUser(username, password string) (string, *model.CredentialsModel) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &model.CredentialsModel{Username: username, Password: string(hashedPassword)}
	suite.Require().NoError(suite.db.Create(user).Error)

	loginReq := api.Authentication{Username: username, Password: password}
	w := suite.performRequest("POST", "/login", loginReq, "")
	suite.Require().Equal(http.StatusOK, w.Code)
	var resp api.AuthenticationResponse
	suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	return resp.Token, user
}

func boolPtr(b bool) *bool {
	return &b
}

func (suite *HandlerIntegrationSuite) TestLoginUser() {
	suite.getAuthTokenForTestUser("test@user.com", "password123")

	suite.Run("Failure - Non-existent user", func() {
		req := api.Authentication{Username: "nouser@example.com", Password: "password"}
		w := suite.performRequest("POST", "/login", req, "")
		suite.Equal(http.StatusUnauthorized, w.Code)
	})
}

func (suite *HandlerIntegrationSuite) TestFormAndStepsCRUD() {
	// 1. Setup: Get a token for an authenticated user
	token, _ := suite.getAuthTokenForTestUser("crud@user.com", "password123")

	var createdFormID string
	var createdStepID string

	// 2. Test Create Form
	suite.Run("Create Form", func() {
		seqReq := api.FormCreate{
			Name:                "CRUD Test Form",
			OpenTrackingEnabled: boolPtr(true),
			Steps: api.FormStepCreateArray{
				{Name: "CRUD Step 1", Content: "First", Step: 1},
			},
		}
		w := suite.performRequest("POST", "/form", seqReq, token)
		suite.Equal(http.StatusCreated, w.Code)
		var resp api.SelfId
		json.Unmarshal(w.Body.Bytes(), &resp)
		createdFormID = resp.Id // Save for next test
	})

	// 3. Test Get Form
	suite.Run("Get Form By ID", func() {
		suite.Require().NotEmpty(createdFormID, "Must have created a form to get it")
		w := suite.performRequest("GET", "/form/"+createdFormID, nil, token)
		suite.Equal(http.StatusOK, w.Code)
		var resp api.FormResponseGet
		json.Unmarshal(w.Body.Bytes(), &resp)
		suite.Equal("CRUD Test Form", resp.Name)
		suite.Equal(1, len(resp.Steps))
		createdStepID = resp.Steps[0].Self.Id
	})

	// 4. Test Update Form
	suite.Run("Update Form By ID", func() {
		suite.Require().NotEmpty(createdFormID, "Must have created a form to update it")
		updateReq := api.FormUpdate{ClickTrackingEnabled: boolPtr(true)}
		w := suite.performRequest("PATCH", "/form/"+createdFormID, updateReq, token)
		suite.Equal(http.StatusOK, w.Code)
		var resp api.FormResponseGet
		json.Unmarshal(w.Body.Bytes(), &resp)
		suite.True(resp.ClickTrackingEnabled)
	})

	// 5. Test Update Form Step
	suite.Run("Update Form Step By ID", func() {
		suite.Require().NotEmpty(createdStepID, "Must have created a step to update it")
		newContent := "Updated Content"
		updateReq := api.FormStepUpdate{Content: &newContent}
		w := suite.performRequest("PATCH", fmt.Sprintf("/form/%s/steps/%s", createdFormID, createdStepID), updateReq, token)
		suite.Equal(http.StatusOK, w.Code)
		var resp api.FormStepResponseGet
		json.Unmarshal(w.Body.Bytes(), &resp)
		suite.Equal("Updated Content", resp.Content)
	})

	// 6. Test Delete Form Step
	suite.Run("Delete Form Step By ID", func() {
		suite.Require().NotEmpty(createdStepID, "Must have created a step to delete it")
		w := suite.performRequest("DELETE", fmt.Sprintf("/form/%s/steps/%s", createdFormID, createdStepID), nil, token)
		suite.Equal(http.StatusNoContent, w.Code)

		wGet := suite.performRequest("GET", fmt.Sprintf("/form/%s/steps/%s", createdFormID, createdStepID), nil, token)
		suite.Equal(http.StatusNotFound, wGet.Code)
	})
}

func (suite *HandlerIntegrationSuite) TestValidationErrors() {
	token, _ := suite.getAuthTokenForTestUser("validator@user.com", "password123")

	suite.Run("Create Form with Invalid Name", func() {
		// Name is empty, which violates `validate:"required"`
		seqReq := api.FormCreate{
			Name:  "",
			Steps: api.FormStepCreateArray{{Name: "s", Content: "c", Step: 1}},
		}
		w := suite.performRequest("POST", "/form", seqReq, token)
		suite.Equal(http.StatusBadRequest, w.Code)

		var resp api.ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		suite.NoError(err)
		suite.Equal(400, resp.Code)
		suite.NotEmpty(resp.Errors)
		suite.Equal("Name", resp.Errors[0].Field)
	})
}

func (suite *HandlerIntegrationSuite) TestNotFound_ForValidButNonExistentUUID() {
	token, _ := suite.getAuthTokenForTestUser("not-found-user@example.com", "password123")

	nonExistentUUID := uuid.New().String()

	suite.Run("GET non-existent form returns 404", func() {
		w := suite.performRequest("GET", "/form/"+nonExistentUUID, nil, token)
		suite.Equal(http.StatusNotFound, w.Code)
	})
}

func (suite *HandlerIntegrationSuite) TestCreateForm_Validation_EmptySteps() {
	token, _ := suite.getAuthTokenForTestUser("validation-user@example.com", "password123")

	suite.Run("Create Form with empty steps array fails validation", func() {
		seqReq := api.FormCreate{
			Name:                "Form with no steps",
			OpenTrackingEnabled: boolPtr(true),
			Steps:               api.FormStepCreateArray{},
		}

		w := suite.performRequest("POST", "/form", seqReq, token)
		suite.Equal(http.StatusBadRequest, w.Code)

		var resp api.ValidationErrorResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		suite.Equal(400, resp.Code)

		found := false
		for _, detail := range resp.Errors {
			if detail.Field == "Steps" {
				found = true
				suite.Contains(detail.Message, "min")
			}
		}
		suite.True(found, "Expected a validation error for the 'Steps' field")
	})
}
