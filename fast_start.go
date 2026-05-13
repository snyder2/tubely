package main

import "os/exec"

func processVideoForFastStart(filePath string) (string, error) {
	outputPath := filePath + ".processing"

	process := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)

	err := process.Run()
	if err != nil {
		return "", err
	}

	return outputPath, nil
}
