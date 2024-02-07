FROM golang:alpine as builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o /videostreamsegmentschecker

FROM busybox
EXPOSE 8081
COPY --from=builder /videostreamsegmentschecker /home
WORKDIR /home
ENTRYPOINT [ "./videostreamsegmentschecker" ]