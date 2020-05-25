#!/bin/sh

CHROOT_USER="trusted-cgi"

systemctl stop trusted-cgi.service || echo "failed to stop service"
systemctl disable trusted-cgi.service || echo "failed to disable service"

if id -u ${CHROOT_USER}; then
  echo "Removing user ${CHROOT_USER}..."
  userdel -r ${CHROOT_USER}
fi


