package s3

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
	"github.com/dharmab/hyperboard/internal/storage"
	"github.com/rs/zerolog/log"
)

// Storage implements storage.Storage using an S3-compatible object store.
type Storage struct {
	client   *s3.Client
	bucket   string
	endpoint string
}

// New creates a new S3 Storage.
func New(ctx context.Context, endpoint, bucket, accessKey, secretKey, region string, usePathStyle bool) (*Storage, error) {
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

	st := &Storage{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
	}

	if err := st.ensureBucketExists(ctx); err != nil {
		return nil, fmt.Errorf("ensure bucket %q exists: %w", bucket, err)
	}

	return st, nil
}

func (st *Storage) ensureBucketExists(ctx context.Context) error {
	_, err := st.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(st.bucket),
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

	log.Info().Str("bucket", st.bucket).Msg("bucket does not exist, creating")
	_, err = st.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(st.bucket),
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

// Ping checks connectivity to the S3-compatible object store.
func (st *Storage) Ping(ctx context.Context) error {
	_, err := st.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(st.bucket),
	})
	if err != nil {
		return fmt.Errorf("head bucket %q: %w", st.bucket, err)
	}
	return nil
}

// Upload uploads data to the given key and returns the public URL.
func (st *Storage) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	_, err := st.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(st.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload %s: %w", key, err)
	}
	url := fmt.Sprintf("%s/%s/%s", st.endpoint, st.bucket, key)
	return url, nil
}

// Download retrieves an object by key.
func (st *Storage) Download(ctx context.Context, key string) (*storage.Object, error) {
	out, err := st.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(st.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", key, err)
	}
	ct := "application/octet-stream"
	if out.ContentType != nil {
		ct = *out.ContentType
	}
	var cl int64
	if out.ContentLength != nil {
		cl = *out.ContentLength
	}
	return &storage.Object{Body: out.Body, ContentType: ct, ContentLength: cl}, nil
}

// Delete removes an object at the given key.
func (st *Storage) Delete(ctx context.Context, key string) error {
	_, err := st.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(st.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete %s: %w", key, err)
	}
	return nil
}
