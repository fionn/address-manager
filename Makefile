SRC :- $(git ls-files *.go)

bin/fb_mock: $(SRC) go.mod go.sum
	go build -v -o $@ github.com/fionn/address-manager/cmd/$(@F)

bin/service: $(SRC) go.mod go.sum
	go build -v -o $@ github.com/fionn/address-manager/cmd/$(@F)
