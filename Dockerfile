FROM golang

#get source code
RUN git clone https://github.com/coveo/uabot-server.git


WORKDIR /go/uabot-server/
RUN go get -d 

RUN rm /go/src/github.com/coveo/uabot/scenariolib/visit.go
RUN mv /go/uabot-server/hack/visit.go /go/src/github.com/coveo/uabot/scenariolib/visit.go



#run server
CMD [ "go", "run", "main.go", "-queue-length=20", "-port=5000", "-routinesPerCPU=3", "-silent=true"]

