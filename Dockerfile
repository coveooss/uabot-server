FROM golang

#get source code
RUN git clone https://github.com/coveo/uabot-server.git
RUN rm src/github.com/coveo/uabot/scenariolib/visit.go
RUN mv /go/uabot-server/visit.go src/github.com/coveo/uabot/scenariolib/visit.go



WORKDIR /go/uabot-server/
RUN go get -d

#run server
CMD [ "go", "run", "main.go", "-queue-length=20", "-port=5000", "-routinesPerCPU=3", "-silent=true"]

