# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/github.com/replicatedhq/gatekeeper
COPY pkg/ pkg/
COPY cmd/ cmd/
COPY vendor/ vendor/

# Build Manager
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/replicatedhq/gatekeeper/cmd/manager

# Build CLI
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o gatekeeper github.com/replicatedhq/gatekeeper/cmd/gatekeeper

# Copy the controller-manager into a thin image
FROM ubuntu:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/replicatedhq/gatekeeper/manager .
COPY --from=builder /go/src/github.com/replicatedhq/gatekeeper/gatekeeper .
ENTRYPOINT ["./manager"]
