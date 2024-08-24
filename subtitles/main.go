package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "cloud.google.com/go/speech/apiv1/speechpb"
	translate "cloud.google.com/go/translate/apiv3"
	translatepb "cloud.google.com/go/translate/apiv3/translatepb"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	projectID := os.Getenv("PROJECT_ID")
	sourceLanguage := os.Getenv("SOURCE_LANGUAGE")
	targetLanguage := os.Getenv("TARGET_LANGUAGE")

	speechClient, err := speech.NewClient(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Fatalf("Failed to create speech client: %v", err)
	}
	defer speechClient.Close()

	translateClient, err := translate.NewTranslationClient(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Fatalf("Failed to create translate client: %v", err)
	}
	defer translateClient.Close()

	var subtitles []string
	processedSegments := make(map[string]bool)

	for {
		segmentAudio, err := getLatestSegment()
		if err != nil {
			log.Printf("Error getting latest segment: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if len(segmentAudio) == 0 {
			log.Println("No audio content in the latest segment, skipping")
			time.Sleep(5 * time.Second)
			continue
		}

		segmentHash := fmt.Sprintf("%x", md5.Sum(segmentAudio))
		if processedSegments[segmentHash] {
			log.Println("Segment already processed, skipping")
			time.Sleep(5 * time.Second)
			continue
		}

		translatedText, err := transcribeAndTranslate(ctx, speechClient, translateClient, segmentAudio, projectID, sourceLanguage, targetLanguage)
		if err != nil {
			log.Printf("Error in transcription/translation: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		subtitles = append(subtitles, translatedText)
		processedSegments[segmentHash] = true

		err = createWebVTT(subtitles)
		if err != nil {
			log.Printf("Error creating WebVTT: %v", err)
		}

		time.Sleep(5 * time.Second)
	}
}

func transcribeAndTranslate(ctx context.Context, speechClient *speech.Client, translateClient *translate.TranslationClient, audioContent []byte, projectID string, sourceLanguage string, targetLanguage string) (string, error) {
	if len(audioContent) == 0 {
		return "", fmt.Errorf("audio content is empty")
	}

	req := &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_LINEAR16,
			SampleRateHertz: 16000,
			LanguageCode:    sourceLanguage,
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: audioContent},
		},
	}

	resp, err := speechClient.Recognize(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to recognize speech: %v", err)
	}

	var fullTranscript string
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			fullTranscript += alt.Transcript + " "
		}
	}

	if fullTranscript == "" {
		return "", fmt.Errorf("no recognition result returned")
	}

	translateResp, err := translateClient.TranslateText(ctx, &translatepb.TranslateTextRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/global", projectID),
		Contents:           []string{fullTranscript},
		TargetLanguageCode: targetLanguage,
	})
	if err != nil {
		return "", fmt.Errorf("failed to translate: %v", err)
	}

	if len(translateResp.Translations) == 0 {
		return "", fmt.Errorf("no translation returned")
	}

	return translateResp.Translations[0].TranslatedText, nil
}

func createWebVTT(subtitles []string) error {
	vttContent := "WEBVTT\n\n"
	for i, subtitle := range subtitles {
		startTime := i * 10
		endTime := startTime + 10
		vttContent += fmt.Sprintf("%s --> %s\n%s\n\n", formatTime(startTime), formatTime(endTime), subtitle)
	}

	return os.WriteFile("../stream_hls/subtitles.vtt", []byte(vttContent), 0644)
}

func formatTime(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d.000", h, m, s)
}

func getLatestSegment() ([]byte, error) {
	manifestFile, err := os.Open("../stream_hls/stream.m3u8")
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest file: %v", err)
	}
	defer manifestFile.Close()

	scanner := bufio.NewScanner(manifestFile)
	var latestSegment string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, ".ts") {
			latestSegment = line
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading manifest file: %v", err)
	}

	if latestSegment == "" {
		return nil, fmt.Errorf("no .ts files found in manifest")
	}

	segmentPath := filepath.Join("../stream_hls", latestSegment)
	log.Printf("Processing segment: %s", segmentPath)

	fileInfo, err := os.Stat(segmentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for segment %s: %v", segmentPath, err)
	}
	log.Printf("Segment file size: %d bytes", fileInfo.Size())

	file, err := os.Open(segmentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open segment file %s: %v", segmentPath, err)
	}
	defer file.Close()

	header := make([]byte, 188)
	_, err = file.Read(header)
	if err != nil {
		return nil, fmt.Errorf("failed to read segment file header %s: %v", segmentPath, err)
	}
	if header[0] != 0x47 {
		return nil, fmt.Errorf("invalid MPEG-TS file: %s", segmentPath)
	}

	tempFile, err := os.CreateTemp("", "audio_*.raw")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	cmd := exec.Command("ffmpeg", "-i", segmentPath, "-vn", "-acodec", "pcm_s16le", "-f", "s16le", "-ac", "1", "-ar", "16000", "-y", tempFile.Name())
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to extract audio from segment %s: %v", latestSegment, err)
	}

	fileInfo, err = os.Stat(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for extracted audio: %v", err)
	}
	log.Printf("Extracted audio file size: %d bytes", fileInfo.Size())

	segmentAudio, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read temp audio file: %v", err)
	}

	if len(segmentAudio) == 0 {
		return nil, fmt.Errorf("extracted audio is empty for segment: %s", latestSegment)
	}

	log.Printf("Successfully extracted %d bytes of audio from segment: %s", len(segmentAudio), latestSegment)

	return segmentAudio, nil
}
