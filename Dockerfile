FROM golang:latest AS build
LABEL maintainer="Afshin Paydar <afshinpaydar@gmail.com>"
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/kubectl-deploy .
FROM bitnami/kubectl:latest
COPY --from=build /out/kubectl-deploy /bin/kubectl-deploy