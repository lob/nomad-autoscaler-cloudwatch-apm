LDFLAGS := "-s -w"
export CGO_ENABLED=0

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
	GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -o dist/nomad-autoscaler-cloudwatch-apm_linux_amd64
	GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -o dist/nomad-autoscaler-cloudwatch-apm_linux_arm64
	GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -o dist/nomad-autoscaler-cloudwatch-apm_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -o dist/nomad-autoscaler-cloudwatch-apm_darwin_arm64
	GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -o dist/nomad-autoscaler-cloudwatch-apm_windows_amd64

