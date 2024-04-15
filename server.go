package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	pb "github.com/rammyblog/file-upload-grpc/protos"
	"google.golang.org/grpc"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"time"
)

func generateRandomFileNameString(length int, filePath string) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	ext := path.Ext(filePath)
	return string(b) + ext
}

type Server struct {
	pb.UnimplementedFileServiceServer
	uploader *manager.Uploader
}

func NewServer() *Server {

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(s3Client)

	return &Server{
		uploader: uploader,
	}
}

func (s Server) WriteToPipe(pw *io.PipeWriter, content []byte) error {
	if _, err := pw.Write(content); err != nil {
		err = pw.CloseWithError(err)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s Server) Upload(stream pb.FileService_UploadServer) error {
	req, err := stream.Recv() // First, receive to get the filename
	if err != nil {
		return err
	}
	randomFileName := generateRandomFileNameString(30, req.Filename)
	pr, pw := io.Pipe()
	done := make(chan error)

	go func() {
		defer func(pw *io.PipeWriter) {
			err := pw.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(pw)
		_, err := s.uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("AWS_BUCKET_NAME")), // Specify your S3 bucket name
			Key:    aws.String(randomFileName),
			Body:   pr,
		})
		done <- err
	}()
	err = s.WriteToPipe(pw, req.Content)
	if err != nil {
		return err
	}
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// close pipe
			err := pw.Close()
			if err != nil {
				return err
			}
			break
		}
		if err != nil {
			err := pw.CloseWithError(err)
			if err != nil {
				return err
			}
		}
		err = s.WriteToPipe(pw, req.Content)
		if err != nil {
			return err
		}

	}

	// Wait for the upload to complete
	if err := <-done; err != nil {
		return err
	}
	return stream.SendAndClose(&pb.FileUploadSuccess{
		Success: true,
		Message: "Upload success",
	})
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pb.RegisterFileServiceServer(grpcServer, NewServer())
	log.Println("Server listening at", lis.Addr())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
