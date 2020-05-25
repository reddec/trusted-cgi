#!/bin/sh

CHROOT_USER="trusted-cgi"

if ! id -u ${CHROOT_USER}; then
  echo "Creating chroot user ${CHROOT_USER}..."
  useradd -M -c "trusted cgi dummy user" -r -s /bin/nologin ${CHROOT_USER}
fi

systemctl enable trusted-cgi.service || echo "failed to enable service"
systemctl start trusted-cgi.service || echo "failed to start service"
