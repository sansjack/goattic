package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"goattic/client/internal/compress"
	"goattic/client/internal/config"
	"goattic/client/internal/ffmpeg"

	"github.com/gen2brain/beeep"
)

var videoExts = map[string]bool{
	".mov": true,
	".mp4": true,
}

var imageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

const defaultAPIURL = "https://api.jacksansom.com"

type uploadResponse struct {
	PresignedPost struct {
		URL    string            `json:"url"`
		Fields map[string]string `json:"fields"`
	} `json:"presignedPost"`
	Key       string `json:"key"`
	PublicURL string `json:"publicUrl"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage:\n  goattic configure\n  goattic config:list\n  goattic <file>\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "configure":
		runConfigure()
		return
	case "config":
		runConfigList()
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	filePath := os.Args[1]

	info, err := os.Stat(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is a directory\n", filePath)
		os.Exit(1)
	}

	filePath, err = maybeCompress(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compressing file: %v\n", err)
		os.Exit(1)
	}

	filename := filepath.Base(filePath)
	fmt.Printf("Uploading %s...\n", filename)

	upload, err := requestUploadURL(cfg, filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting upload URL: %v\n", err)
		os.Exit(1)
	}

	if err := uploadToS3(upload, filePath); err != nil {
		fmt.Fprintf(os.Stderr, "Error uploading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Uploaded! Public URL (available after validation):\n%s\n", upload.PublicURL)
	_ = beeep.Notify("GoAttic Upload", fmt.Sprintf("Uploaded %s", filename), "")
}

func runConfigure() {
	cfg := &config.Config{APIURL: defaultAPIURL}

	fmt.Printf("API URL [%s]: ", defaultAPIURL)
	var apiURL string
	fmt.Scanln(&apiURL)
	if apiURL != "" {
		cfg.APIURL = apiURL
	}

	fmt.Print("API Key: ")
	var apiKey string
	fmt.Scanln(&apiKey)
	cfg.APIKey = apiKey

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Config saved.")
}

func runConfigList() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("apiUrl:  %s\n", cfg.APIURL)
	fmt.Printf("apiKey:  %s\n", cfg.APIKey)
}

func maybeCompress(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	if imageExts[ext] {
		out := filePath[:len(filePath)-len(ext)] + "_compressed" + ext
		origInfo, _ := os.Stat(filePath)
		ok, err := compress.Image(filePath, out, ext)
		if err != nil {
			return filePath, fmt.Errorf("image compress: %w", err)
		}
		if ok {
			compInfo, _ := os.Stat(out)
			fmt.Printf("Compressed: %dKB → %dKB\n", origInfo.Size()/1024, compInfo.Size()/1024)
			return out, nil
		}
		return filePath, nil
	}

	if !videoExts[ext] {
		return filePath, nil
	}

	ffmpegBin, err := ffmpeg.BinaryPath()
	if err != nil {
		return filePath, nil //no ffmpeg binary found- skipping compression
	}

	out := filePath[:len(filePath)-len(ext)] + "_compressed" + ext
	fmt.Printf("Compressing video...\n")

	cmd := exec.Command(ffmpegBin, "-y", "-i", filePath,
		"-vcodec", "libx264", "-crf", "28", "-preset", "fast",
		"-acodec", "aac", "-b:a", "128k",
		out,
	)
	if err := cmd.Run(); err != nil {
		return filePath, fmt.Errorf("ffmpeg: %w", err)
	}

	origInfo, _ := os.Stat(filePath)
	compInfo, _ := os.Stat(out)
	if compInfo.Size() >= origInfo.Size() {
		os.Remove(out)
		return filePath, nil
	}

	fmt.Printf("Compressed: %dMB → %dMB\n", origInfo.Size()/1024/1024, compInfo.Size()/1024/1024)
	return out, nil
}

func requestUploadURL(cfg *config.Config, filename string) (*uploadResponse, error) {
	body, _ := json.Marshal(map[string]string{"filename": filename})

	req, err := http.NewRequest("POST", cfg.APIURL+"/upload", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", cfg.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, respBody)
	}

	var upload uploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&upload); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &upload, nil
}

func uploadToS3(upload *uploadResponse, filePath string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for k, v := range upload.PresignedPost.Fields {
		if err := writer.WriteField(k, v); err != nil {
			return fmt.Errorf("failed to write field %s: %w", k, err)
		}
	}

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	writer.Close()

	resp, err := http.Post(upload.PresignedPost.URL, writer.FormDataContentType(), &buf)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("S3 returned %d: %s", resp.StatusCode, respBody)
	}

	return nil
}
