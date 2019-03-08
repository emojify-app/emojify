FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

RUN mkdir /service

COPY ./emojify-service /service/
COPY ./images /service/images/

WORKDIR /service

ENTRYPOINT /service/emojify-service
