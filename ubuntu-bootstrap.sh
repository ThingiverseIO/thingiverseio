#!/usr/bin/env bash

sudo apt-get update -y
sudo apt-get install -y curl make git libtool build-essential dh-autoreconf pkg-config mercurial dh-autoreconf

curl -o ./zeromq.tar.gz http://download.zeromq.org/zeromq-4.0.4.tar.gz
tar -C . -zxvf ./zeromq.tar.gz
rm ./zeromq.tar.gz
cd zeromq-4.0.4
./autogen.sh
./configure
make
sudo make install
sudo ldconfig
cd ..
sudo ifconfig

wget https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz
tar -xf go1.6.linux-amd64.tar.gz

export GOROOT=/home/vagrant/go
echo 'export GOROOT=/home/vagrant/go' > /home/vagrant/.bashrc
export GOPATH=/home/vagrant/goprog
echo 'export GOPATH=/home/vagrant/goprog' > /home/vagrant/.bashrc
export PATH=/home/vagrant/go/bin:$PATH
echo 'export PATH=/home/vagrant/go/bin:$PATH' > /home/vagrant/.bashrc

go get github.com/nu7hatch/gouuid
go get gopkg.in/gcfg.v1
go get github.com/joernweissenborn/eventual2go
go get gopkg.in/yaml.v2
go get github.com/pebbe/zmq4
go get github.com/hashicorp/memberlist
go get github.com/ugorji/go/codec

ln -s /vagrant goprog/src/github.com/joernweissenborn/thingiverseio
