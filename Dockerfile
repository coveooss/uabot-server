FROM golang

#get source code
RUN git clone https://github.com/coveo/uabot-server.git

WORKDIR /go/uabot-server/
RUN go get -d

EXPOSE 8443

#run server
CMD [ "go", "run", "main.go", "-queue-length=20", "-sslport=8443", "-routinesPerCPU=3", "-silent=false"]
