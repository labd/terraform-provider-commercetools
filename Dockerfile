FROM golang:1.13.1-stretch AS build-env
WORKDIR /terraform-provider

ADD . /terraform-provider

ENV GOPROXY=https://proxy.golang.org
RUN go mod download
RUN go build -o terraform-provider-commercetools

# final stage
FROM hashicorp/terraform:0.12.9

RUN apk add libc6-compat

WORKDIR /config

COPY --from=build-env /terraform-provider/terraform-provider-commercetools /bin
