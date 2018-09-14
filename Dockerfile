FROM golang:1.11.0-stretch AS build-env
WORKDIR /terraform-provider

ADD . /terraform-provider

RUN go mod download
RUN go build -o terraform-provider-commercetools

# final stage
FROM hashicorp/terraform:light

WORKDIR /config

COPY --from=build-env /terraform-provider/terraform-provider-commercetools /bin
