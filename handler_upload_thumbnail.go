package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	const maxMemory = 10485760

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, 500, "Failed to parse multipart form", err)
	}

	fileData, fileHeader, err := r.FormFile("thumbnail")
	mediaType := fileHeader.Header.Get("Content-Type")

	parsedType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, 500, "Failed to parse media type", err)
	}

	if parsedType != "image/jpeg" && parsedType != "image/png" {
		respondWithError(w, 500, "Incorrect media type", errors.New("Incorrect media type"))
	}

	parsedTypeSplit := strings.Split(parsedType, "/")

	filePath := filepath.Join(cfg.assetsRoot, videoIDString+"."+parsedTypeSplit[1])
	file, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, 500, "Failed to create thumbnail file", err)
	}

	_, err = io.Copy(file, fileData)
	if err != nil {
		respondWithError(w, 500, "Failed to copy thumbnail file data", err)
	}

	// imageData, err := io.ReadAll(fileData)
	// if err != nil {
	// 	respondWithError(w, 500, "Failed to read thumbnail data", err)
	// 	return
	// }

	// dataString := base64.StdEncoding.EncodeToString(imageData)
	dataURL := "http://localhost:" + cfg.port + "/assets/" + videoIDString + "." + parsedTypeSplit[1]

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, 500, "Failed to fetch video", err)
		return
	}

	if videoData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
		return
	}

	videoData.ThumbnailURL = &dataURL
	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, 500, "Failed to update video in database", err)
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
