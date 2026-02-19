// TODO: move to shared package as API can tap into this too!
package validate

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const headerSize = 12

type fileSignature struct {
	Name   string
	Magic  []byte
	Offset int
}

var allowList = []fileSignature{
	{Name: "jpeg", Magic: []byte{0xFF, 0xD8, 0xFF}},
	{Name: "png", Magic: []byte{0x89, 0x50, 0x4E, 0x47}},
	{Name: "gif", Magic: []byte{0x47, 0x49, 0x46, 0x38}},
	{Name: "webp", Magic: []byte{0x52, 0x49, 0x46, 0x46}},
	{Name: "webp", Magic: []byte{0x57, 0x45, 0x42, 0x50}, Offset: 8},
	{Name: "mp4", Magic: []byte{0x66, 0x74, 0x79, 0x70}, Offset: 4},
	{Name: "mov", Magic: []byte{0x71, 0x74, 0x20, 0x20}, Offset: 8}, // QuickTime: ftyp brand "qt  "
}

var mimeTypes = map[string]string{
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"webp": "image/webp",
	"mp4":  "video/mp4",
	"mov":  "video/quicktime",
}

func DetectFileType(header []byte) (string, bool) {
	for _, sig := range allowList {
		end := sig.Offset + len(sig.Magic)
		if len(header) < end {
			continue
		}
		if bytes.Equal(header[sig.Offset:end], sig.Magic) {
			return sig.Name, true
		}
	}
	return "", false
}

func ReadHeader(ctx context.Context, client *s3.Client, bucket, key string) ([]byte, error) {
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Range:  aws.String(fmt.Sprintf("bytes=0-%d", headerSize-1)),
	})
	if err != nil {
		return nil, fmt.Errorf("read header from s3://%s/%s: %w", bucket, key, err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func MimeType(fileType string) string {
	if mt, ok := mimeTypes[fileType]; ok {
		return mt
	}
	return "application/octet-stream"
}

func CopyObject(ctx context.Context, client *s3.Client, srcBucket, destBucket, key, fileType string) error {
	_, err := client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:            aws.String(destBucket),
		CopySource:        aws.String(fmt.Sprintf("%s/%s", srcBucket, key)),
		Key:               aws.String(key),
		ContentType:       aws.String(MimeType(fileType)),
		MetadataDirective: types.MetadataDirectiveReplace,
	})
	if err != nil {
		return fmt.Errorf("copy s3://%s/%s → s3://%s/%s: %w", srcBucket, key, destBucket, key, err)
	}
	return nil
}

func DeleteObject(ctx context.Context, client *s3.Client, bucket, key string) error {
	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete s3://%s/%s: %w", bucket, key, err)
	}
	return nil
}
