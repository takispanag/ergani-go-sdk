FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./

RUN go mod tidy

COPY . .

RUN go build -o ergani-app .

CMD ["./ergani-app"]