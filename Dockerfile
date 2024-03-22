FROM golang:1.21.5-alpine AS build

WORKDIR /app

COPY . .

RUN go mod download

WORKDIR /app/cmd

RUN go build -o project-service

FROM busybox:latest

WORKDIR /project-service

COPY --from=build /app/cmd/project-service .

COPY --from=build /app/cmd/.env .

EXPOSE 50002

CMD [ "./project-service" ]