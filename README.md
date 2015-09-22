# goard


### setup

```
mkdir goard
cd goard
export GOPATH=`pwd`
go get github.com/argami/goard
cd src/github.com/argami/goard
go install

export PATH=$PATH:$GOPATH/bin

./goard &
```

### Adding remote (temp)

/:servername/remote/add/?addr=https://remote_url:8443&password=pass

### Listing containers

/:servername/list_containers
