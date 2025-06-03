package db

// Package db provides database access and model operations.
//
// This file (ext.go) contains extension methods for database models.
// These extension methods include:
//   - Data transformations between models and DTOs
//   - Helper functions to extract specific data from models
//   - Utility methods to convert models to different formats/representations
//   - Business logic that extends the base model functionality
//
// Extension methods should be kept separate from core model definitions
// to maintain clean separation of concerns.

import (
	"errors"
	"net/url"
	"path"
	"strings"
)

// S3BucketInfo contains parsed information from an S3 URL
type S3BucketInfo struct {
	BucketName string // Name of the S3 bucket
	Path       string // Full path without the filename
	Filename   string // Just the filename
}

// GetBucketInfo extracts bucket name, path, and filename from an S3 URL
// Example URL: https://starburst-webvitals-data-n-virginia.s3.us-east-1.amazonaws.com/snapshots/snapshots_https_3A_2F_2Fwww.google.com_2F_ss-7bTBaMu1RK_2025-05-05T12-38-57-945Z/metrics.json.gz
func GetBucketInfo(s3URL string) (S3BucketInfo, error) {
	if s3URL == "" {
		return S3BucketInfo{}, errors.New("s3URL is empty")
	}

	// Parse the URL
	parsedURL, err := url.Parse(s3URL)
	if err != nil {
		return S3BucketInfo{}, err
	}

	// Extract bucket name from the hostname
	// Format: bucket-name.s3.region.amazonaws.com
	hostnameParts := strings.Split(parsedURL.Hostname(), ".")
	bucketName := hostnameParts[0]

	// Get the path without leading slash
	fullPath := strings.TrimPrefix(parsedURL.Path, "/")

	// Split path into directory and filename
	dir, filename := path.Split(fullPath)

	return S3BucketInfo{
		BucketName: bucketName,
		Path:       dir,
		Filename:   filename,
	}, nil
}

func (m Snapshot) GetBucketInfo() (S3BucketInfo, error) {
	return GetBucketInfo(m.FullReportPath.String)
}
