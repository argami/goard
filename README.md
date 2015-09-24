# goard

### NOTES

apt-get install libprotobuf-dev libprotobuf-c0-dev protobuf-c-compiler protobuf-compiler python-protobuf
http://criu.org/Installation

### live migrate

https://insights.ubuntu.com/2015/05/06/live-migration-in-lxd/

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
