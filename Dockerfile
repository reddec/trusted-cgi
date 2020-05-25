FROM alpine
RUN apk add --no-cache python3 py3-setuptools py3-virtualenv php nodejs npm make && pip3 install requests && npm install -g axios
EXPOSE 3434
VOLUME /data
WORKDIR /data
ENV INITIAL_ADMIN_PASSWORD admin
ENV BIND 0.0.0.0:3434
COPY trusted-cgi /
ENTRYPOINT ["/trusted-cgi", "--disable-chroot"]