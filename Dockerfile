FROM alpine
ADD ./drone-webhook /bin/
RUN apk -Uuv add ca-certificates
CMD /bin/drone-webhook
