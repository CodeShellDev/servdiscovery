FROM alpine:3.22
RUN apk --no-cache add ca-certificates

ARG IMAGE_TAG
ENV IMAGE_TAG=$IMAGE_TAG
LABEL org.opencontainers.image.version=$IMAGE_TAG

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .

COPY dist/${TARGETOS}/${TARGETARCH}/app .

RUN rm dist/ -r

CMD ["./app"]