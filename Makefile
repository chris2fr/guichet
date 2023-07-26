BIN=guichet
SRC=main.go model.go view.go controller.go utils.go model-user.go view-admin.go view-home.go view-invite.go view-login.go view-passwd.go view.profile.go utils-http.go utils-ldap.go utils-config.go directory.go garage.go picture.go session.go utils-ssha.go 

# ssha.go profile.go admin.go invite.go directory.go utils.go picture.go login.go config.go http-utils.go home.go model-user.go gpas.go session.go model.go view.go controller.go utils-ldap.go

DOCKER=lxpz/guichet_amd64

all: $(BIN)

$(BIN): $(SRC)
	go get -d -v
	go build -v -o $(BIN)

$(BIN).static: $(SRC)
	go get -d -v
	CGO_ENABLED=0 GOOS=linux go build -a -v -o $(BIN).static

docker: $(BIN).static
	docker build -t $(DOCKER):$(TAG) .
	docker push $(DOCKER):$(TAG)
	docker tag $(DOCKER):$(TAG) $(DOCKER):latest
	docker push $(DOCKER):latest
