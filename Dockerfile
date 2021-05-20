FROM golang:alpine AS builder
WORKDIR $GOPATH/src/slurm-sre-auth-service/
COPY . .
RUN go mod tidy -v \
    && go mod vendor \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -v -o /go/bin/auth-service main.go

FROM scratch
ENTRYPOINT ["/app/auth-service"]
EXPOSE 2121
COPY --from=builder /go/bin/auth-service /app/auth-service