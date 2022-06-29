FROM golang:1.18-alpine

RUN mkdir -p /app

WORKDIR /app

COPY . /app/

RUN go mod download all

# COPY *.go ./

RUN cd cmd && go build -o /go-app

# CMD ["/go-app"]

COPY waitformysql.sh /waitformysql.sh
RUN chmod +x /waitformysql.sh
RUN apk add --no-cache bash
CMD [ "/waitformysql.sh","mysqldb:3306", "--", "/go-app" ]
