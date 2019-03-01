FROM alpine
ADD ./drone-webhook /bin/
RUN chmod 755 /bin/drone-webhook \
    && apk -Uuv add ca-certificates
ENTRYPOINT /bin/drone-webhook
