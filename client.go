package main

import (
	"context"
	"fmt"
	pb "github.com/rammyblog/file-upload-grpc/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"os"
)

func uploadFile(client pb.FileServiceClient, filePath string) {
	stream, err := client.Upload(context.Background())
	if err != nil {
		log.Fatalf("%v.Upload(_) = _, %v", client, err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("os.Open(_) = _, %v", filePath)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("file.Close(_) = %v", err)
		}
	}(file)
	buf := make([]byte, 1024)

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("file.Read(_) = _, %v", err)
		}
		err = stream.Send(&pb.FileUploadRequest{
			Filename: filePath,
			Content:  buf[:n],
		})
		if err != nil {
			return
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got %v, want %v", stream, err, nil)
	}
	log.Printf("Upload result: %v", res)
	log.Printf("Upload finished: %v", res.Message)
}

func main() {

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("could not close connection: %v", err)
		}
	}(conn)

	client := pb.NewFileServiceClient(conn)

	if len(os.Args) < 3 {
		fmt.Println("Usage: [upload/download] [file name]")
		return
	}
	operation, filePath := os.Args[1], os.Args[2]

	switch operation {
	case "upload":
		uploadFile(client, filePath)
	}
}
