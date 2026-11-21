#!/bin/sh
sed -i 's/^AllowTcpForwarding no/AllowTcpForwarding yes/' /config/sshd/sshd_config
