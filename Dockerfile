FROM nimlang/nim:1.6.10-alpine
RUN apk add --no-cache python3 py3-setuptools py3-virtualenv php nodejs npm make git
EXPOSE 3434
VOLUME /data
WORKDIR /data
ENV INITIAL_ADMIN_PASSWORD admin
ENV BIND 0.0.0.0:3434
COPY trusted-cgi /
ENTRYPOINT ["/trusted-cgi", "--disable-chroot"]