#!/bin/sh
set -eu

# Squid runs as 'squid' user but needs to write logs to stdout/stderr
chmod 666 /dev/stdout /dev/stderr 2>/dev/null || true

# Squid spawns squid-helper as the 'squid' user, so it needs write access
chown squid:squid /var/lib/squid-brocker

exec squid -N -f /etc/squid/squid.conf
