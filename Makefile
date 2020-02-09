all: guichet

guichet: main.go
	go get -v
	go build -v
