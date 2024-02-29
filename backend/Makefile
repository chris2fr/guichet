BIN=guichet
SRC=models/model.go models/user.go models/passwd.go views/admin.go views/home.go views/invite.go views/login.go views/passwd.go views/user.go views/http.go models/ldap.go models/config.go views/directory.go  views/picture.go views/session.go utils/ssha.go models/modelutils.go views/view.go controllers/controller.go models/authentik.go utils/files.go main.go

#SRC=models/model.go models/user.go models/passwd.go views/admin.go views/home.go views/invite.go views/login.go views/passwd.go views/user.go utils/http.go utils/ldap.go utils/config.go  utils/ssha.go utils/lesutils.go views/view.go controller.go main.go

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
