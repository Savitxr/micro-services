# FROM golang:1.26.1-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 AS builder
# WORKDIR /src
# COPY go.mod go.sum ./
# RUN go mod download
# COPY . .
# FROM gcr.io/distroless/static 
# WORKDIR /src
# COPY --from=builder /productcatalogservice ./server
# COPY products.json .
# EXPOSE 3550
# ENTRYPOINT ["/src/server"]

FROM golang:1.26.1-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 AS builder

WORKDIR /src

# Install protoc
RUN apk add --no-cache protoc

# Install Go protobuf plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy Go module files
COPY ./services/product-catalog/go.mod ./services/product-catalog/go.sum ./

# Fix dependencies (this is the key part)
RUN go get github.com/lib/pq && go mod tidy

# Copy protos and product-catalog source
COPY ./protos ./protos
COPY ./services/product-catalog/ ./

# Generate proto files
RUN PATH=$PATH:$(go env GOPATH)/bin && protoc --proto_path=./protos --go_out=./genproto --go_opt=paths=source_relative --go-grpc_out=./genproto --go-grpc_opt=paths=source_relative ./protos/demo.proto

# Build the binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /productcatalogservice .

# Copy products.json for the final stage
COPY ./services/product-catalog/products.json .

# Final minimal image
FROM gcr.io/distroless/static

WORKDIR /src

COPY --from=builder /productcatalogservice ./server
COPY --from=builder /src/products.json .

EXPOSE 3550

ENTRYPOINT ["/src/server"]