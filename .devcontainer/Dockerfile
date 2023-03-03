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

# 1. 安装protobuf编译器: protoc v22.0  -> https://github.com/protocolbuffers/protobuf/releases -> protoc-22.0-linux-x86_64.zip
# 2. 安装生成go code插件: protoc-gen-go v1.28.0 -> https://github.com/protocolbuffers/protobuf-go/releases -> protoc-gen-go.v1.28.0.linux.amd64.tar.gz
# 3. 安装生成grpc code插件: protoc-gen-grpc v1.3.0 -> https://github.com/grpc/grpc-go/releases ->  protoc-gen-go-grpc.v1.3.0.linux.amd64.tar.gz
# 4. 安装生成grpc-gateway code插件: protoc-gen-grpc-gateway v2.15.0 -> https://github.com/grpc-ecosystem/grpc-gateway/releases -> protoc-gen-grpc-gateway-v2.15.0-linux-x86_64
# 5. 安装生成swagger文档code插件: protoc-gen-openapiv2 v2.15.0 -> https://github.com/grpc-ecosystem/grpc-gateway/releases -> protoc-gen-openapiv2-v2.15.0-linux-x86_64
 
 RUN wget -P /tmp/ https://github.com/protocolbuffers/protobuf/releases/download/v22.0/protoc-22.0-linux-x86_64.zip \
 && unzip -o  /tmp/protoc-22.0-linux-x86_64.zip -d /tmp/protoc-22.0-linux-x86_64 \
 && cp /tmp/protoc-22.0-linux-x86_64/bin/protoc /usr/local/bin \
 && mkdir /usr/local/include \
 && cp -p /tmp/protoc-22.0-linux-x86_64/include/google /usr/local/include

 RUN wget -P /tmp/ https://github.com/protocolbuffers/protobuf-go/releases/download/v1.28.0/protoc-gen-go.v1.28.0.linux.amd64.tar.gz \
 && tar -xvzf /tmp/protoc-gen-go.v1.28.0.linux.amd64.tar.gz -C . \
 && cp /tmp/protoc-gen-go /usr/local/bin

 RUN wget -P /tmp/ https://github.com/grpc/grpc-go/releases/download/cmd%2Fprotoc-gen-go-grpc%2Fv1.3.0/protoc-gen-go-grpc.v1.3.0.linux.amd64.tar.gz \
 && mkdir /tmp/protoc-gen-go-grpc \
 && tar -xvzf /tmp/protoc-gen-go-grpc.v1.3.0.linux.amd64.tar.gz -C /tmp/protoc-gen-go-grpc \
 && cp /tmp/protoc-gen-go-grpc/protoc-gen-go-grpc /usr/local/bin

RUN wget -P /tmp/ https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.15.0/protoc-gen-grpc-gateway-v2.15.0-linux-x86_64 \
 && cp /tmp/protoc-gen-grpc-gateway-v2.15.0-linux-x86_64 /tmp/protoc-gen-grpc-gateway \
 && chmod +x /tmp/protoc-gen-grpc-gateway \
 && mv /tmp/protoc-gen-grpc-gateway /usr/local/bin

RUN wget -P /tmp/ https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v2.15.0/protoc-gen-openapiv2-v2.15.0-linux-x86_64 \
 && cp /tmp/protoc-gen-openapiv2-v2.15.0-linux-x86_64 /tmp/protoc-gen-openapiv2 \
 && chmod +x /tmp/protoc-gen-openapiv2 \
 && mv /tmp/protoc-gen-openapiv2 /usr/local/bin

# 安装protoc接口基础类型定义
RUN git -C /usr/local/include clone https://github.com/googleapis/googleapis.git \
&& git -C /usr/local/include clone https://github.com/grpc-ecosystem/grpc-gateway.git

ENV PROTO_PATH=/usr/local/include

WORKDIR ${WORKSPACE}

