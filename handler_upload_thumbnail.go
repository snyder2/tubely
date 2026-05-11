package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10485760

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, 500, "Failed to parse multipart form", err)
	}

	fileData, fileHeader, err := r.FormFile("thumbnail")
	mediaType := fileHeader.Header.Get("Content-Type")

	imageData, err := io.ReadAll(fileData)
	if err != nil {
		respondWithError(w, 500, "Failed to read thumbnail data", err)
		return
	}

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, 500, "Failed to fetch video", err)
		return
	}

	if videoData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
		return
	}

	thumbnailData := thumbnail{
		data:      imageData,
		mediaType: mediaType,
	}

	videoThumbnails[videoID] = thumbnailData

	newURL := "http://localhost:8091/api/thumbnails/" + videoIDString

	videoData.ThumbnailURL = &newURL
	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, 500, "Failed to update video in database", err)
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
