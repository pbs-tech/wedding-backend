all: build_auth build_rsvp

build_auth:
	GOARCH=arm64 GOOS=linux go build -tags lambda.norpc -o ./bin/bootstrap ./auth/handler.go
	(cd bin && zip -FS auth.zip bootstrap)
	rm -rf ./bin/bootstrap

build_rsvp:
	GOARCH=arm64 GOOS=linux go build -tags lambda.norpc -o ./bin/bootstrap ./rsvp/handler.go
	(cd bin && zip -FS rsvp.zip bootstrap)
	rm -rf ./bin/bootstrap