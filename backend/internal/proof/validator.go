package proof

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// ValidationResult holds the result of proof validation
type ValidationResult struct {
	Valid       bool
	ImageHash   string
	TakenAt     *time.Time // nil if EXIF not available
	Errors      []string
	Warnings    []string
}

// ValidatePhotoProof validates a photo proof submission
// - Decodes base64 image
// - Extracts EXIF data to verify photo timestamp
// - Generates SHA256 hash for duplicate detection
func ValidatePhotoProof(imageBase64 string, challengeStart, challengeEnd time.Time) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Remove data URL prefix if present
	imageData := imageBase64
	if idx := strings.Index(imageBase64, ","); idx != -1 {
		imageData = imageBase64[idx+1:]
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 image: %w", err)
	}

	if len(decoded) == 0 {
		return nil, errors.New("empty image data")
	}

	// Generate SHA256 hash
	hash := sha256.Sum256(decoded)
	result.ImageHash = fmt.Sprintf("%x", hash)

	// Try to extract EXIF data
	takenAt, exifErr := extractPhotoTime(decoded)
	if exifErr != nil {
		result.Warnings = append(result.Warnings, "EXIF 데이터를 읽을 수 없습니다")
	} else if takenAt != nil {
		result.TakenAt = takenAt

		// Validate photo was taken within challenge period
		// Allow 1 day buffer before start and after end for timezone issues
		startWithBuffer := challengeStart.Add(-24 * time.Hour)
		endWithBuffer := challengeEnd.Add(24 * time.Hour)

		if takenAt.Before(startWithBuffer) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("사진이 챌린지 시작 전에 촬영되었습니다 (촬영: %s)", takenAt.Format("2006-01-02")))
		}
		if takenAt.After(endWithBuffer) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("사진이 챌린지 종료 후에 촬영되었습니다 (촬영: %s)", takenAt.Format("2006-01-02")))
		}
	}

	return result, nil
}

// extractPhotoTime extracts the original photo timestamp from EXIF data
func extractPhotoTime(imageData []byte) (*time.Time, error) {
	reader := bytes.NewReader(imageData)
	x, err := exif.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("decode exif: %w", err)
	}

	// Try DateTimeOriginal first (when photo was actually taken)
	dt, err := x.DateTime()
	if err != nil {
		return nil, fmt.Errorf("no datetime in exif: %w", err)
	}

	return &dt, nil
}

// ValidateStepsProof validates a steps proof submission
// For steps, we trust the provided hash as verification is done client-side
func ValidateStepsProof(stepsHash string) (*ValidationResult, error) {
	if strings.TrimSpace(stepsHash) == "" {
		return nil, errors.New("steps hash is required")
	}

	return &ValidationResult{
		Valid:     true,
		ImageHash: stepsHash,
		Errors:    []string{},
		Warnings:  []string{},
	}, nil
}

// HashImage generates SHA256 hash from raw image bytes
func HashImage(r io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", fmt.Errorf("hash image: %w", err)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// HashBase64Image generates SHA256 hash from base64 encoded image
func HashBase64Image(imageBase64 string) (string, error) {
	// Remove data URL prefix if present
	imageData := imageBase64
	if idx := strings.Index(imageBase64, ","); idx != -1 {
		imageData = imageBase64[idx+1:]
	}

	decoded, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	hash := sha256.Sum256(decoded)
	return fmt.Sprintf("%x", hash), nil
}
