FROM alpine
EXPOSE 3434
VOLUME /data
WORKDIR /data
ENV INITIAL_ADMIN_PASSWORD admin
COPY trusted-cgi /
ENTRYPOINT ["/trusted-cgi", "--disable-chroot"]