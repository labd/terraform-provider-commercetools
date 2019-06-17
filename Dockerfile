FROM golang:1.12.1-stretch AS build-env
WORKDIR /terraform-provider

ADD . /terraform-provider

RUN go mod download
RUN go build -o terraform-provider-commercetools

# final stage
FROM hashicorp/terraform:0.12.2

WORKDIR /config

COPY --from=build-env /terraform-provider/terraform-provider-commercetools /bin
