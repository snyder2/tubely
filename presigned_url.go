package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	client := s3.NewPresignClient(s3Client)
	presignedReq, err := client.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}

	return presignedReq.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	url := *video.VideoURL
	splitURL := strings.Split(url, ",")
	fmt.Println("Split URL:", splitURL)

	if len(splitURL) != 2 {
		return video, nil
	}

	presignedURL, err := generatePresignedURL(cfg.s3Client, splitURL[0], splitURL[1], time.Second*10)
	if err != nil {
		return video, err
	}

	video.VideoURL = &presignedURL

	return video, nil
}
