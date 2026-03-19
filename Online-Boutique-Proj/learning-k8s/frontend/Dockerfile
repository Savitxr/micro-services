FROM golang:1.26.1-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY ../protos ./protos
COPY . .
# Generate protobuf files (fixes undefined types like pb.Order)
RUN sed -i 's|protodir=../../protos|protodir=./protos|g' genproto.sh
RUN chmod +x genproto.sh && ./genproto.sh
# Remove mustMapEnv calls for missing services to avoid panics on unset env vars
RUN sed -i '/mustMapEnv.*CURRENCY_SERVICE_ADDR/d' main.go && \
    sed -i '/mustMapEnv.*RECOMMENDATION_SERVICE_ADDR/d' main.go && \
    sed -i '/mustMapEnv.*CHECKOUT_SERVICE_ADDR/d' main.go && \
    sed -i '/mustMapEnv.*SHIPPING_SERVICE_ADDR/d' main.go && \
    sed -i '/mustMapEnv.*AD_SERVICE_ADDR/d' main.go && \
    sed -i '/mustMapEnv.*SHOPPING_ASSISTANT_SERVICE_ADDR/d' main.go
# Remove mustConnGRPC calls for missing services to avoid connection panics
RUN sed -i '/mustConnGRPC.*currencySvcConn/d' main.go && \
    sed -i '/mustConnGRPC.*recommendationSvcConn/d' main.go && \
    sed -i '/mustConnGRPC.*checkoutSvcConn/d' main.go && \
    sed -i '/mustConnGRPC.*shippingSvcConn/d' main.go && \
    sed -i '/mustConnGRPC.*adSvcConn/d' main.go
# Remove unused imports if they persist (adjust based on exact errors)
RUN sed -i '/"math\/rand"/d' handlers.go && \
    sed -i '/"time"/d' handlers.go
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /go/bin/frontend .
FROM gcr.io/distroless/static
WORKDIR /src
COPY --from=builder /go/bin/frontend /src/server
COPY ./templates ./templates
COPY ./static ./static
EXPOSE 8080
ENTRYPOINT ["/src/server"]
