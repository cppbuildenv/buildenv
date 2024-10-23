package main

import (
	"buildenv/cmd"
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	if exit := cmd.Listen(); exit {
		return
	}

	upload()
	fmt.Println("----hello world")
}

func upload() {
	ctx := context.Background()
	endpoint := "192.168.100.27:9000"
	accessKeyID := "PMh3Ghxu4KXP8fK94mVj"
	secretAccessKey := "uiUYkJFXXNNmrnnXMO8is0mm5mICGdbszJjx1h0V"
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	bucketName := "buildenv"
	// location := "platform"

	// if err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location}); err != nil {
	// 	// Check to see if we already own this bucket (which happens if you run this twice)
	// 	exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
	// 	if errBucketExists == nil && exists {
	// 		log.Printf("We already own %s\n", bucketName)
	// 	} else {
	// 		log.Fatalln(err)
	// 	}
	// } else {
	// 	log.Printf("Successfully created %s\n", bucketName)
	// }

	// Upload the test file
	// Change the value of filePath if the file is in another location
	objectName := "Clash.for.Windows-0.20.39-win.7z"
	filePath := "D:/Software/Clash.for.Windows-0.20.39-win.7z"
	contentType := "application/octet-stream"

	// Upload the test file with FPutObject
	info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
}
