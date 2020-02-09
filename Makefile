all: guichet

guichet: main.go ssha.go profile.go admin.go
	go get -v
	go build -v
