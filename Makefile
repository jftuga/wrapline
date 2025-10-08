
wrapline: wrapline.go
	go build -ldflags="-s -w"

test: wrapline
	go test -v

clean:
	command rm -f wrapline cpu*.prof .??*~ .DS_Store

