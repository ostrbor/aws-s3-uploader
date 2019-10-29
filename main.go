package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
)

func main() {
	var bucket, key string
	var file *os.File
	defer file.Close()
	// TODO get bucket, key and file from rabbitmq

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	result, err := svc.ListBuckets(nil)
	if err != nil {
		fmt.Println("Unable to list buckets")
		os.Exit(1)
	}
	var found bool
	for _, b := range result.Buckets {
		bucketName := aws.StringValue(b.Name)
		if bucketName == bucket {
			found = true
		}
	}
	if !found {
		_, err = svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			fmt.Println("Unable to create bucket")
			os.Exit(1)
		}
	}

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
