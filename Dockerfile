FROM golang:1.13.8 

LABEL maintainer="Adrian Hynes"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o detector .

EXPOSE 2112

CMD ["./detector"]