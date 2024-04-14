package main

import (
	pb "github.com/rammyblog/file-upload-grpc/protos"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
)

type server struct {
	pb.UnimplementedFileServiceServer
}

func (s server) Upload(stream pb.FileService_UploadServer) error {
	var fileName string
	var fileData []byte

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			outFile, err := os.Create(fileName)
			if err != nil {
				return err
			}
			defer func(outFile *os.File) {
				err := outFile.Close()
				if err != nil {
					log.Fatal(err)
				}
			}(outFile)
			if _, err := outFile.Write(fileData); err != nil {
				return err
			}
			return stream.SendAndClose(&pb.FileUploadSuccess{
				Success: true,
				Message: "Upload success",
			})
		}
		if err != nil {
			return err
		}
		fileName = "file_ser.proto"
		fileData = append(fileData, req.Content...)

	}

}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterFileServiceServer(grpcServer, new(server))
	log.Println("Server listening at", lis.Addr())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
