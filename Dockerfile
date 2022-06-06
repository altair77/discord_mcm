FROM golang:1.17.8-buster

ENV JAVA_HOME=/opt/jdk-18.0.1.1
ENV PATH=$PATH:$JAVA_HOME/bin

RUN curl -O https://download.java.net/java/GA/jdk18.0.1.1/65ae32619e2f40f3a9af3af1851d6e19/2/GPL/openjdk-18.0.1.1_linux-x64_bin.tar.gz && \
    tar zxf openjdk-18.0.1.1_linux-x64_bin.tar.gz -C /opt && \
    rm openjdk-18.0.1.1_linux-x64_bin.tar.gz

WORKDIR /go/src/app

COPY *.go go.mod go.sum ./

RUN go mod download

CMD ["go", "run", "."]
