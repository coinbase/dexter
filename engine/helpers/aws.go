package helpers

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

//
// Defines a local directory to use as the backend for Dexter.
// Overrides all calls to S3, used for demo purposes.
//
var LocalDemoPath = ""

//
// Build the directory structure for a local Dexter demo.
//
func BuildDemoPath() {
	if LocalDemoPath == "" {
		return
	}
	if !strings.HasSuffix(LocalDemoPath, "/") {
		LocalDemoPath = LocalDemoPath + "/"
	}
	err := os.MkdirAll(filepath.FromSlash(LocalDemoPath), 0777)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "helpers.BuildDemoPath",
			"error": err.Error(),
		}).Fatal("unable to build local demo path")
	}
	err = os.MkdirAll(filepath.FromSlash(LocalDemoPath+"investigations"), 0777)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "helpers.BuildDemoPath",
			"error": err.Error(),
		}).Fatal("unable to build local demo path investigations directory")
	}
	err = os.MkdirAll(filepath.FromSlash(LocalDemoPath+"reports"), 0777)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "helpers.BuildDemoPath",
			"error": err.Error(),
		}).Fatal("unable to build local demo path reports directory")
	}
	err = os.MkdirAll(filepath.FromSlash(LocalDemoPath+"investigators"), 0777)
	if err != nil {
		log.WithFields(log.Fields{
			"at":    "helpers.BuildDemoPath",
			"error": err.Error(),
		}).Fatal("unable to build local demo path investigators directory")
	}
}

//
// Download a file from the Dexter S3 bucket.
//
func GetS3File(name string) ([]byte, error) {
	if LocalDemoPath != "" {
		return ioutil.ReadFile(LocalDemoPath + name)
	}
	svc := s3.New(session.New())
	input := &s3.GetObjectInput{
		Bucket: S3Bucket(),
		Key:    aws.String(name),
	}

	result, err := svc.GetObject(input)
	if err != nil {
		return []byte{}, err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(result.Body)
	if err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}

//
// List the contents of a path in the Dexter S3 bucket.
//
func ListS3Path(path string) ([]string, error) {
	if LocalDemoPath != "" {
		infos, err := ioutil.ReadDir(LocalDemoPath + path)
		set := []string{}
		for _, info := range infos {
			set = append(set, path+info.Name())
		}
		return set, err
	}
	svc := s3.New(session.New())
	input := &s3.ListObjectsInput{
		Bucket:  S3Bucket(),
		MaxKeys: aws.Int64(1000),
		Prefix:  aws.String(path),
	}

	result, err := svc.ListObjects(input)
	if err != nil {
		return []string{}, err
	}

	strs := []string{}
	for _, object := range result.Contents {
		strs = append(strs, *object.Key)
	}
	return strs, nil
}

//
// Upload data to a file in the Dexter S3 bucket.
//
func UploadS3File(path string, data io.ReadSeeker) error {
	if LocalDemoPath != "" {
		bytes, _ := ioutil.ReadAll(data)
		return ioutil.WriteFile(LocalDemoPath+path, bytes, 0644)
	}
	_, err := s3.New(session.New()).PutObject(&s3.PutObjectInput{
		ACL:                  aws.String(s3.ObjectCannedACLBucketOwnerFullControl),
		Body:                 data,
		Bucket:               S3Bucket(),
		Key:                  aws.String(path),
		ServerSideEncryption: aws.String("AES256"),
	})
	return err
}

//
// Delete a file from the Dexter S3 bucket.
//
func DeleteS3File(path string) error {
	if LocalDemoPath != "" {
		return os.Remove(filepath.FromSlash(LocalDemoPath + path))
	}
	_, err := s3.New(session.New()).DeleteObject(&s3.DeleteObjectInput{
		Bucket: S3Bucket(),
		Key:    aws.String(path),
	})
	return err
}
