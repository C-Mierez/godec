package api

import (
	"context"
	"time"
)

type MediaHandlers struct{}

func NewMediaHandlers() *MediaHandlers {
	return &MediaHandlers{}
}

func (h *MediaHandlers) GetMediaUploadURL(ctx context.Context, request GetMediaUploadURLRequestObject) (GetMediaUploadURLResponseObject, error) {
	// TODO: Replace with actual S3 presigned URL generation
	uploadURL := "https://storage.example.com/upload/" + time.Now().Format("20060102150405")

	return GetMediaUploadURL200JSONResponse(UploadURLResponse{
		UploadURL: uploadURL,
	}), nil
}
