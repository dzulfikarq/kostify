package minio

import (
	"context"
	"fmt"
	"io"
	"strings"

	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Uploader struct {
	client    *miniogo.Client
	bucket    string
	secure    bool
	publicURL string
}

func NewUploader(endpoint, accessKey, secretKey, bucket string, secure bool, publicURL string) (*Uploader, error) {
	client, err := miniogo.New(endpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	return &Uploader{client: client, bucket: bucket, secure: secure, publicURL: publicURL}, nil
}

func (u *Uploader) EnsureBucket(ctx context.Context) error {
	exists, err := u.client.BucketExists(ctx, u.bucket)
	if err != nil {
		return fmt.Errorf("cek bucket %s: %w", u.bucket, err)
	}
	if exists {
		return nil
	}
	if err := u.client.MakeBucket(ctx, u.bucket, miniogo.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("buat bucket %s: %w", u.bucket, err)
	}
	return nil
}

func (u *Uploader) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := u.client.PutObject(ctx, u.bucket, key, r, size, miniogo.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("upload object %s: %w", key, err)
	}
	return nil
}

func (u *Uploader) Delete(ctx context.Context, key string) error {
	return u.client.RemoveObject(ctx, u.bucket, key, miniogo.RemoveObjectOptions{})
}

func (u *Uploader) PublicURL(key string) string {
	if u.publicURL != "" {
		return strings.TrimSuffix(u.publicURL, "/") + "/" + u.bucket + "/" + key
	}
	scheme := "https"
	if !u.secure {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, u.client.EndpointURL().Host, u.bucket, key)
}
