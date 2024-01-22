
FROM golang:1.21 as builder
WORKDIR /app
COPY . /app
RUN go build -o myapp
RUN apt-get update && apt-get install -y postgresql-client
COPY wait-for-db.sh /wait-for-db.sh
RUN chmod +x /wait-for-db.sh
CMD ["/wait-for-db.sh"]
