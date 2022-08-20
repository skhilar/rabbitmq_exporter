FROM golang:1.18-alpine3.16 AS builder
WORKDIR /workspace
# Update and install packages
RUN apk update && \
    apk add git make bash gcc musl-dev
# Copy over files
COPY . .
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo,staticbuild -ldflags "-X 'main.Version=v1.0.0' -X 'main.Revision=dev'  \
    -X 'main.Branch=master' -X 'main.BuildDate=2022-08-20'" -a -o ./rabbitmq_exporter

FROM alpine:3.16
# Install packages
RUN apk --update add --no-cache ca-certificates
# Copy over built exporter
COPY --from=builder /workspace/rabbitmq_exporter /rabbitmq_exporter
# Declare the port on which the webserver will be exposed.
# As we're going to run the executable as an unprivileged user, we can't bind
# to ports below 1024.
EXPOSE 9419
# Perform any further action as an unprivileged user.
USER 65535:65535
# Check if exporter is alive; 10 retries gives prometheus some time to retrieve bad data (5 minutes)
HEALTHCHECK --retries=10 CMD ["/rabbitmq_exporter", "-check-url", "http://localhost:9419/health"]
# Run the compiled binary.
ENTRYPOINT ["/rabbitmq_exporter"]

