FROM golang:1.16.0-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /kubectl-deploy .
FROM bitnami/kubectl:latest
COPY --from=build /kubectl-deploy /opt/bitnami/kubectl/bin/kubectl-deploy
