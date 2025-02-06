package blobstore

import (
	"context"
	"errors"
	"net/http"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3BlobStore struct {
	client *s3.Client
}

func NewS3BlobStore(client *s3.Client) *S3BlobStore {
	return &S3BlobStore{client: client}
}

func (s *S3BlobStore) Put(ctx context.Context, bucket string, key string, blob Blob) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        blob.Body,
		ContentType: blob.ContentType,
		Metadata:    blob.Metadata,
	})

	if err != nil {
		var httpResponseErr *awshttp.ResponseError
		if errors.As(err, &httpResponseErr) &&
			(httpResponseErr.HTTPStatusCode() == http.StatusNotFound ||
				httpResponseErr.HTTPStatusCode() == http.StatusForbidden) {
			return nil
		}
		return err
	}

	return nil
}

func (s *S3BlobStore) Exist(ctx context.Context, bucket string, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		var httpResponseErr *awshttp.ResponseError
		if errors.As(err, &httpResponseErr) &&
			(httpResponseErr.HTTPStatusCode() == http.StatusNotFound ||
				httpResponseErr.HTTPStatusCode() == http.StatusForbidden) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *S3BlobStore) Get(ctx context.Context, bucket string, key string) (Blob, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		// If the object that you request doesn’t exist, the error that Amazon S3 returns
		// depends on whether you also have the s3:ListBucket permission.
		// If you have the s3:ListBucket permission on the bucket, Amazon S3 returns an HTTP status code 404 (Not Found) error.
		// If you don’t have the s3:ListBucket permission, Amazon S3 returns an HTTP status code 403 ("access denied") error.
		var httpResponseErr *awshttp.ResponseError
		if errors.As(err, &httpResponseErr) &&
			(httpResponseErr.HTTPStatusCode() == http.StatusNotFound ||
				httpResponseErr.HTTPStatusCode() == http.StatusForbidden) {
			return Blob{}, nil
		}
		return Blob{}, err
	}

	return Blob{
		Body:        output.Body,
		ContentType: output.ContentType,
		Metadata:    output.Metadata,
	}, nil
}
