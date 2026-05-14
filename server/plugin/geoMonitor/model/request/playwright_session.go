package request

type StartPlaywrightSessionRequest struct {
	PlatformID uint `json:"platformId" binding:"required"`
}
