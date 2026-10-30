package schemas

// Health is the response body for the health route.
type Health struct {
	Status string `json:"status" validate:"required" example:"ok"`
}
