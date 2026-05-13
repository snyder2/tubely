package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	http.MaxBytesReader(w, r.Body, 1073741824)

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

	fmt.Println("uploading video", videoID, "by user", userID)

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, 500, "Failed to fetch video", err)
		return
	}

	if videoData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
		return
	}

	fileData, fileHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, 500, "Failed to create multipart file from video data", err)
	}
	defer fileData.Close()

	mediaType := fileHeader.Header.Get("Content-Type")

	parsedType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, 500, "Failed to parse media type", err)
	}

	if parsedType != "video/mp4" {
		respondWithError(w, 500, "Incorrect media type", errors.New("Incorrect media type"))
	}

	videoFile, err := os.CreateTemp("", "video_upload.mp4")
	if err != nil {
		respondWithError(w, 500, "Failed to create temporary video file", err)
	}

	defer os.Remove(videoFile.Name())
	defer videoFile.Close()

	_, err = io.Copy(videoFile, fileData)
	if err != nil {
		respondWithError(w, 500, "Failed to copy video file to disk", err)
	}

	aspectRatio, err := getVideoAspectRatio(videoFile.Name())
	if err != nil {
		log.Printf("Error: %s", err)
		respondWithError(w, 500, "Failed to get video aspect ratio", err)
	}

	// fmt.Println("aspect ratio:", aspectRatio)

	var keyPrefix string
	switch aspectRatio {
	case "16:9":
		keyPrefix = "landscape"
	case "9:16":
		keyPrefix = "portrait"
	default:
		keyPrefix = "other"
	}

	processedVideoPath, err := processVideoForFastStart(videoFile.Name())
	if err != nil {
		respondWithError(w, 500, "Failed to process video for fast start", err)
	}

	processedVideo, err := os.Open(processedVideoPath)
	if err != nil {
		respondWithError(w, 500, "Failed to open processed video", err)
	}

	_, err = videoFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, 500, "Failed to reset temp file pointer", err)
	}

	randomData := make([]byte, 32)
	rand.Read(randomData)
	urlSequence := base64.RawURLEncoding.EncodeToString(randomData)

	key := keyPrefix + "/" + urlSequence + ".mp4"

	putObjectParams := &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &key,
		Body:        processedVideo,
		ContentType: &parsedType,
	}

	_, err = cfg.s3Client.PutObject(r.Context(), putObjectParams)
	if err != nil {
		respondWithError(w, 500, "Failed to put object to s3 bucket", err)
	}

	videoURL := "https://" + cfg.s3Bucket + ".s3." + cfg.s3Region + ".amazonaws.com/" + key
	videoData.VideoURL = &videoURL

	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, 500, "Failed to update video in database", err)
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
