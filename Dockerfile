#Dockerfile for operator-sdk builder image
FROM golang:1.11

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
  mkdir -p $GOPATH/src/github.com/operator-framework && \
  cd $GOPATH/src/github.com/operator-framework && \
  git clone https://github.com/operator-framework/operator-sdk && \
 cd operator-sdk && \
 git checkout v0.5.0 && \
 make dep && \
 make install

 RUN apt-get update && apt-get install apt-transport-https dirmngr -y
 RUN echo 'deb https://apt.dockerproject.org/repo debian-stretch main' >> /etc/apt/sources.list && apt-get update && apt-get install docker-engine -y --allow-unauthenticated

