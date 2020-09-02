package api

import (
	"encoding/json"
	"fmt"
)

// the root of the request object
type GenesisTemplateRequest struct {
	Template GenesisTemplate `json:"genesis_template"`
}

type GenesisTemplate struct {
	Name     string   `json:"name"`
	Language string   `json:"language"`
	Runtime  string   `json:"runtime"`
	Options  []Option `json:"options"`
}

type Option struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type GenesisPayload struct {
	UserID        string        `json:"userID"`
	ProjectName   string        `json:"projectName"`
	TemplateName  string        `json:"templateName"`
	Options       []Option      `json:"options"`
	JenkinsUrl    string        `json:"jenkinsUrl"`
	EnableWebhook bool          `json:"enableWebhook"`
	TargetRepo    BitBucketRepo `json:"targetRepo"`
}

type BitBucketRepo struct {
	ProjectKey       string `json:"projectKey"`
	ProjectDomain    string `json:"projectDomain"`
	RepositorySlug   string `json:"repositorySlug"`
	FunctionalDomain string `json:"functionalDomain"`
	ProjectName      string `json:"projectName"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewErrorResponse(code int, msg string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: msg,
	}
}

func NewErrorResponseJson(code int, msg string) []byte {
	errorMessage := NewErrorResponse(code, msg)
	errorPayload, err := json.Marshal(errorMessage)

	if err != nil {
		fmt.Printf("something happened while marshalling error message: %+v", err)
	}

	return errorPayload
}

type SuccessResponse struct {
	Code    int         `json:"code"`
	Payload interface{} `json:"payload"`
}

func NewSuccessResponse(code int, payload interface{}) SuccessResponse {
	return SuccessResponse{
		Code:    code,
		Payload: payload,
	}
}

func NewSuccessResponseJson(code int, payload interface{}) []byte {
	successMessage := NewSuccessResponse(code, payload)
	successPayload, err := json.Marshal(successMessage)

	if err != nil {
		fmt.Printf("something happened while marshalling success message: %+v", err)
	}

	return successPayload
}
