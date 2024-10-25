all: build_auth build_rsvp

build_auth:
	GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -tags lambda.norpc -o ./bin/bootstrap ./auth/main.go
	(cd bin && zip -FS auth.zip bootstrap)

build_rsvp:
	GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -tags lambda.norpc -o ./bin/bootstrap ./rsvp/main.go
	(cd bin && zip -FS rsvp.zip bootstrap)
