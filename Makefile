LDFLAGS := "-s -w"
.PHONY: all

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -ldflags $(LDFLAGS) -a -o dist/nomad-autoscaler-cloudwatch-apm

.PHONY: dist
dist:
	rm -rf dist
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -o "dist/nomad-autoscaler-cloudwatch-apm_linux_amd64"
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags $(LDFLAGS) -a -o "dist/nomad-autoscaler-cloudwatch-apm_linux_arm"
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -o "dist/nomad-autoscaler-cloudwatch-apm_darwin_amd64"
