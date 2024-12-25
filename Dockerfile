FROM golang:1.23-alpine
WORKDIR /app
COPY . .
RUN go build -o app .
EXPOSE 8000
CMD ["./app"]

