all: build_auth build_rsvp

build_auth:
	GOOS=linux GOARCH=amd64 go build -o ./bin/auth ./auth/handler.go
	zip -j ./bin/auth.zip ./bin/auth

build_rsvp:
	GOOS=linux GOARCH=amd64 go build -o ./bin/rsvp ./rsvp/handler.go
	zip -j ./bin/rsvp.zip ./bin/rsvpgo 