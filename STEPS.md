go mod init github.com/jcsawyer123/simple-go-api

<!-- Get depedencies -->
go get github.com/go-chi/chi/v5
go get github.com/sony/gobreaker
go get github.com/aws/aws-sdk-go-v2/service/sns
go get github.com/aws/aws-sdk-go-v2/service/sqs
go get github.com/go-resty/resty/v2


<!-- Generic Project Structure -->
project/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   │   └── auth.go
│   ├── config/
│   │   └── config.go
│   ├── handlers/
│   │   └── handlers.go
│   └── server/
│       └── server.go
├── go.mod
└── go.sum


<!-- Run -->
go run cmd/server/main.go
