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
	"github.com/aws/smithy-go"
	"github.com/dharmab/hyperboard/internal/storage"
	"github.com/rs/zerolog/log"
)

// Storage implements storage.MediaStore using an S3-compatible object store.
type Storage struct {
	client   *s3.Client
	bucket   string
	endpoint string
	region   string
}

// New creates a new S3 Storage.
func New(ctx context.Context, endpoint, bucket, accessKey, secretKey, region string, usePathStyle bool) (*Storage, error) {
	if region == "" {
		region = "us-east-1"
	}
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
		region:   region,
	}

	if err := st.ensureBucketExists(ctx); err != nil {
		return nil, fmt.Errorf("ensure bucket %q exists: %w", bucket, err)
	}

	return st, nil
}

// ensureBucketExists creates the configured S3 bucket if it does not already exist.
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
	input := &s3.CreateBucketInput{Bucket: aws.String(st.bucket)}
	if st.region != "" && st.region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(st.region),
		}
	}
	_, err = st.client.CreateBucket(ctx, input)
	if err == nil {
		return nil
	}
	if !isBucketCreationConflict(err) {
		return fmt.Errorf("create bucket: %w", err)
	}

	// Another instance may have created the bucket after our initial check.
	// Only accept the conflict after confirming that this client can access it.
	if _, headErr := st.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(st.bucket)}); headErr != nil {
		verificationErr := fmt.Errorf("subsequent check: %w", headErr)
		return fmt.Errorf("create bucket conflict: %w", errors.Join(err, verificationErr))
	}
	return nil
}

// isNotFoundError checks whether an error represents an HTTP 404 response.
func isNotFoundError(err error) bool {
	// Fallback check for providers that return untyped 404 errors.
	var respErr interface{ HTTPStatusCode() int }
	if errors.As(err, &respErr) && respErr.HTTPStatusCode() == 404 {
		return true
	}
	return false
}

func isBucketCreationConflict(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.ErrorCode() == "BucketAlreadyOwnedByYou" || apiErr.ErrorCode() == "BucketAlreadyExists"
}

func (st *Storage) objectError(ctx context.Context, operation, key string, err error) error {
	if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
		return fmt.Errorf("%s %s: %w", operation, key, errors.Join(storage.ErrNotFound, err))
	}

	if _, ok := errors.AsType[*types.NotFound](err); ok {
		// HeadObject commonly returns the same generic 404 for a missing key and
		// a missing bucket. Confirm the bucket before classifying the key as absent.
		if _, headErr := st.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(st.bucket)}); headErr == nil {
			return fmt.Errorf("%s %s: %w", operation, key, errors.Join(storage.ErrNotFound, err))
		}
	}
	return fmt.Errorf("%s %s: %w", operation, key, err)
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

// Size returns the file size of an object in bytes.
func (st *Storage) Size(ctx context.Context, key string) (int64, error) {
	metadata, err := st.Metadata(ctx, key)
	if err != nil {
		return 0, err
	}
	return metadata.ContentLength, nil
}

// Metadata describes an object without retrieving its contents.
func (st *Storage) Metadata(ctx context.Context, key string) (*storage.Metadata, error) {
	out, err := st.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(st.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, st.objectError(ctx, "head object", key, err)
	}
	if out.ContentLength == nil {
		return nil, fmt.Errorf("head object %s: response has no content length", key)
	}
	contentType := "application/octet-stream"
	if out.ContentType != nil {
		contentType = *out.ContentType
	}
	return &storage.Metadata{ContentType: contentType, ContentLength: *out.ContentLength}, nil
}

// Download retrieves an object by key.
func (st *Storage) Download(ctx context.Context, key string) (*storage.Media, error) {
	out, err := st.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(st.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, st.objectError(ctx, "download", key, err)
	}
	ct := "application/octet-stream"
	if out.ContentType != nil {
		ct = *out.ContentType
	}
	var cl int64
	if out.ContentLength != nil {
		cl = *out.ContentLength
	}
	return &storage.Media{
		Body:          out.Body,
		ContentType:   ct,
		ContentLength: cl,
	}, nil
}

// DownloadRange retrieves an inclusive byte range from an object.
func (st *Storage) DownloadRange(ctx context.Context, key string, start, end int64) (*storage.Media, error) {
	out, err := st.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(st.bucket),
		Key:    aws.String(key),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", start, end)),
	})
	if err != nil {
		return nil, st.objectError(ctx, fmt.Sprintf("download range bytes=%d-%d", start, end), key, err)
	}
	contentType := "application/octet-stream"
	if out.ContentType != nil {
		contentType = *out.ContentType
	}
	contentLength := end - start + 1
	if out.ContentLength != nil {
		contentLength = *out.ContentLength
	}
	return &storage.Media{
		Body:          out.Body,
		ContentType:   contentType,
		ContentLength: contentLength,
	}, nil
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
