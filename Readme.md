## Simple s3 AWS file upload built using Go & GRPC

## Usage

Add your application configuration to your .env file in the root of your project:

```
AWS_REGION=
AWS_BUCKET_NAME=
```

- Start the server

```bash
go run server.go
```

- Upload using the client

```bash
go run client.go <filename>
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details