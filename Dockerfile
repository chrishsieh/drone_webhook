FROM alpine
ADD drone-webhook /bin/
RUN apk -Uuv add ca-certificates
ENTRYPOINT /bin/drone-webhook
