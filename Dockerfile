FROM golang

#get source code
RUN git clone https://github.com/coveo/uabot-server.git
WORKDIR /go/uabot-server/
RUN go get -d

#run server
CMD [ "go", "run", "main.go", "-queue-length=20", "-port=5000" ]
