package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"
)

// S3Storage implements Storage using an S3-compatible object store.
type S3Storage struct {
	client   *s3.Client
	bucket   string
	endpoint string
}

// NewS3Storage creates a new S3Storage.
func NewS3Storage(ctx context.Context, endpoint, bucket, accessKey, secretKey, region string, usePathStyle bool) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load S3 config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = usePathStyle
	})

	storage := &S3Storage{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
	}

	if err := storage.ensureBucket(ctx); err != nil {
		return nil, fmt.Errorf("ensure bucket %q exists: %w", bucket, err)
	}

	return storage, nil
}

func (s *S3Storage) ensureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return nil
	}

	var notFound *types.NotFound
	var noSuchBucket *types.NoSuchBucket
	if !errors.As(err, &notFound) && !errors.As(err, &noSuchBucket) {
		// Some providers return a generic 404 smithy error instead of typed errors.
		if !isNotFoundError(err) {
			return fmt.Errorf("check bucket: %w", err)
		}
	}

	log.Info().Str("bucket", s.bucket).Msg("bucket does not exist, creating")
	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}
	return nil
}

func isNotFoundError(err error) bool {
	// Fallback check for providers that return untyped 404 errors.
	var respErr interface{ HTTPStatusCode() int }
	if errors.As(err, &respErr) && respErr.HTTPStatusCode() == 404 {
		return true
	}
	return false
}

// Upload uploads data to the given key and returns the public URL.
func (s *S3Storage) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload %s: %w", key, err)
	}
	url := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
	return url, nil
}

// Download retrieves an object by key.
func (s *S3Storage) Download(ctx context.Context, key string) (*StorageObject, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", key, err)
	}
	ct := "application/octet-stream"
	if out.ContentType != nil {
		ct = *out.ContentType
	}
	return &StorageObject{Body: out.Body, ContentType: ct}, nil
}

// Delete removes an object at the given key.
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete %s: %w", key, err)
	}
	return nil
}
