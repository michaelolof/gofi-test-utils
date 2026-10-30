package schemas

// CreatedAPIKey is the response body for a newly created API key. Secret is
// returned only once.
type CreatedAPIKey struct {
	ID     string `json:"id" validate:"required"`
	Prefix string `json:"prefix" validate:"required"`
	Secret string `json:"secret" validate:"required"`
}
