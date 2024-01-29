# Dockerfile.go

FROM golang:1.21

WORKDIR /

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /

EXPOSE 8080

CMD [ "/awesomeProject" ]