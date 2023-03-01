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
    && yum install -y vim

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
ENV GO111MODULE="on" CGO_ENABLED=0 GOPROXY=https://goproxy.cn,direct GOSUMDB=off PATH=${GOROOT}/bin:${GOPATH}/bin:${HOME}/bin:$PATH

RUN echo ${PATH} \
    && echo ${GOVERSION} \
    && echo ${GO_INSTALL_DIR} \
    && echo ${GOROOT} \
    && echo ${GOPATH} \
    && go version

WORKDIR ${WORKSPACE}

