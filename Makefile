all: build_auth build_refresh

build_auth:
	GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -tags lambda.norpc -o ./bin/bootstrap ./auth/main.go
	(cd bin && zip -FS auth.zip bootstrap)

build_refresh:
	GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -tags lambda.norpc -o ./bin/bootstrap ./refresh/main.go
	(cd bin && zip -FS refresh.zip bootstrap)
