FROM centos:8

LABEL maintainer="dairongpeng dairongpeng@foxmail.com"

ARG BASE_GOPROXY=direct,https://goproxy.cn,https://goproxy.io
ARG HOME=/root
ARG WORKSPACE=${HOME}/workspace
ENV LANG=en_US.UTF-8 LANGUAGE=en_US.UTF-8 LC_COLLATE=C LC_CTYPE=en_US.UTF-8 PS1='[\u@dev \W]\$ '

RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-* \
&& sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*

RUN yum install -y net-tools \
    && yum install -y wget \
    && yum install -y vim \
    && yum install -y unzip \
    && yum install -y zip

# 替换Yum镜像源
RUN mv /etc/yum.repos.d /etc/yum.repos.d.bak && mkdir /etc/yum.repos.d \
&& wget -O /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-8.repo \
&& yum clean all \
&& yum makecache

RUN yum install dlv -y \
    && yum install binutils -y \
    && yum install gdb -y \
    && yum install git -y \
    && git --version

RUN git config --global user.name "dairongpeng" \
&& git config --global user.email "dairongpeng@foxmail.com"

RUN wget -P /tmp/ https://golang.google.cn/dl/go1.18.3.linux-amd64.tar.gz \
    && mkdir -p ${HOME}/go \
    && tar -xvzf /tmp/go1.18.3.linux-amd64.tar.gz -C ${HOME}/go \
    && mv ${HOME}/go/go ${HOME}/go/go1.18.3


ENV GOVERSION=go1.18.3 GO_INSTALL_DIR=${HOME}/go
ENV GOROOT=${GO_INSTALL_DIR}/${GOVERSION} GOPATH=${WORKSPACE}/golang
ENV GO111MODULE="on" CGO_ENABLED=0 GOPROXY=https://goproxy.cn,direct GOSUMDB=off PATH=${GOROOT}/bin:${GOPATH}/bin:${HOME}/bin:${HOME}/local/bin:$PATH

RUN echo ${PATH} \
    && echo ${GOVERSION} \
    && echo ${GO_INSTALL_DIR} \
    && echo ${GOROOT} \
    && echo ${GOPATH} \
    && go version

# protoc v22.0  -> https://github.com/protocolbuffers/protobuf/releases -> protoc-22.0-linux-x86_64.zip
# protoc-gen-go v1.28.0 -> https://github.com/protocolbuffers/protobuf-go/releases -> protoc-gen-go.v1.28.0.darwin.amd64.tar.gz
# protoc-gen-grpc-gateway v2.15.0 -> https://github.com/grpc-ecosystem/grpc-gateway/releases -> protoc-gen-grpc-gateway-v2.15.0-linux-x86_64
# protoc-gen-openapiv2 v2.15.0 -> https://github.com/grpc-ecosystem/grpc-gateway/releases -> protoc-gen-openapiv2-v2.15.0-linux-x86_64
 RUN wget -P /tmp/ https://github.com/protocolbuffers/protobuf/releases/download/v22.0/protoc-22.0-linux-x86_64.zip \
 && unzip -o  /tmp/protoc-22.0-linux-x86_64.zip -d ${HOME}/local/bin

 RUN wget -P /tmp/ https://github.com/protocolbuffers/protobuf-go/releases/download/v1.28.0/protoc-gen-go.v1.28.0.linux.amd64.tar.gz \
 && tar -xvzf /tmp/protoc-gen-go.v1.28.0.linux.amd64.tar.gz -C ${HOME}/local/bin

RUN wget -P /tmp/ https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.15.0/protoc-gen-grpc-gateway-v2.15.0-linux-x86_64 \
 && cp protoc-gen-grpc-gateway-v2.15.0-linux-x86_64 ${HOME}/local/bin \
 && mv ${HOME}/local/bin/protoc-gen-grpc-gateway-v2.15.0-linux-x86_64 protoc-gen-grpc-gateway

RUN wget -P /tmp/ https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.15.0/protoc-gen-openapiv2-v2.15.0-linux-x86_64 \
 && cp protoc-gen-openapiv2-v2.15.0-linux-x86_64 ${HOME}/local/bin \
 && mv ${HOME}/local/bin/protoc-gen-openapiv2-v2.15.0-linux-x86_64 protoc-gen-openapiv2

RUN git clone https://github.com/googleapis/googleapis.git ${HOME}/local/include \
&& git clone https://github.com/grpc-ecosystem/grpc-gateway.git ${HOME}/local/include

ENV PROTO_PATH=${HOME}/local/include

WORKDIR ${WORKSPACE}

