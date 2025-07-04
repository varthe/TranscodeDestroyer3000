FROM golang:1.24 AS builder

WORKDIR /proxy

COPY go.mod ./
RUN go mod download

COPY . .

ARG TARGETARCH
ENV CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH

RUN go build -ldflags="-s -w" -o fmq ./cmd/main.go

FROM gcr.io/distroless/static-debian12 AS runner

WORKDIR /proxy

COPY --from=builder /proxy/fmq /proxy/fmq

EXPOSE 80

ENTRYPOINT ["/proxy/fmq"]
