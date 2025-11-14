FROM golang:1.24 AS build
WORKDIR /app

COPY . .

RUN go build -o server .

FROM gcr.io/distroless/base-debian12
COPY --from=build /app/server /server
EXPOSE 8081
EXPOSE 8082
ENTRYPOINT ["/server"]