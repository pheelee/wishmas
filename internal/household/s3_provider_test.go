package household

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// @test CON-009-AC2 CON-009-AC3 CON-009-AC5 CON-009-AC7 CON-009-AC8
// Opt-in only, for disposable provider containers. This test provisions/deletes
// its own unique bucket; production Wishmas never creates or deletes buckets.
func TestS3CompatibleProvider(t *testing.T) {
	endpoint := os.Getenv("WISHMAS_TEST_S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("requires an explicitly configured disposable S3 provider")
	}
	bucket := os.Getenv("WISHMAS_TEST_S3_BUCKET")
	provision := bucket == ""
	if provision {
		bucket = "wishmas-test-" + time.Now().UTC().Format("20060102150405")
	}
	region := os.Getenv("WISHMAS_TEST_S3_REGION")
	if region == "" {
		region = "us-east-1"
	}
	b, err := NewS3Backend(t.TempDir(), StorageConfig{Endpoint: endpoint, Bucket: bucket, Region: region, AccessKey: os.Getenv("WISHMAS_TEST_S3_ACCESS_KEY_ID"), SecretKey: os.Getenv("WISHMAS_TEST_S3_SECRET_ACCESS_KEY"), Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if provision {
		if _, err = b.client.CreateBucket(context.Background(), &s3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
			t.Fatal("fixture bucket creation failed", err)
		}
		defer b.client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{Bucket: aws.String(bucket)})
	}
	defer b.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String("Fämily/household.yaml")})
	if h, _, _, err := b.Load(); err != nil || h != nil {
		t.Fatal("empty bucket discovery failed", err)
	}
	hash, _ := HashPassword("Fixture123!")
	s, err := b.Create(Household{SchemaVersion: 2, Name: "Fämily", Language: "en-US", AdminPasswordHash: hash, Wishes: []Wish{{ID: "w1", Title: "Provider wish", Picture: "w1.webp", PictureSource: "https://images.example/image.jpg"}}})
	if err != nil {
		t.Fatal("app creation failed", err)
	}
	h, rev, err := s.Load()
	if err != nil || h.Name != "Fämily" {
		t.Fatal("signed read failed", err)
	}
	if err = s.Update(rev, func(h *Household) error { h.Language = "de-CH"; return nil }); err != nil {
		t.Fatal("signed write failed", err)
	}
	h, _, err = s.Load()
	if err != nil || h.Language != "de-CH" {
		t.Fatal("write not retained", err)
	}
	objects, err := b.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	if err != nil || len(objects.Contents) != 1 || aws.ToString(objects.Contents[0].Key) != "Fämily/household.yaml" {
		t.Fatal("unexpected remote objects", err)
	}
	raw, err := os.ReadFile(b.cachePath())
	if err != nil || !strings.Contains(string(raw), "Fämily/household.yaml") {
		t.Fatal("fallback missing", err)
	}
}
