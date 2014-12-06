# -*- encoding: utf-8 -*-

# :enddoc:

#
# Vagrantfile
#
# Copyright 2013-2014 Krzysztof Wilczynski
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# A little helper to determine underlying platform ...
def platform_bits
  ['x'].pack('P').size * 8
end

# Simple provisioning elements ...
script = %{
#!/bin/bash

set -e

export PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

echo 'Starting provisioning ...'

# Avoid the configuration file questions ...
cat <<'EOF' | sudo tee /etc/apt/apt.conf.d/99vagrant &>/dev/null
DPkg::options { "--force-confdef"; "--force-confnew"; }
EOF

# Disable need for user interaction ...
export DEBIAN_FRONTEND=noninteractive

# Update package repositories ...
sudo apt-get update &> /dev/null

# Install needed build time dependencies ...
sudo apt-get install --force-yes -y \
  curl             \
  git              \
  mercurial        \
  make             \
  patch            \
  bison            \
  flex             \
  build-essential  \
  autoconf         \
  automake         \
  autotools-dev    \
  libltdl-dev      \
  libtool          \
  libtool-doc &> /dev/null

# Remove existing "libmagic-dev" package.
sudo apt-get remove --purge --force-yes -y libmagic-dev &> /dev/null

# Clean up unneeded packages ...
{
  sudo apt-get autoremove --force-yes -y
  sudo apt-get autoclean --force-yes -y
  sudo apt-get clean --force-yes -y
} &> /dev/null

# Select appropriate "godeb" package.
# XXX(kwilczynski): Should "gvm" be used instead?
case $(uname -m) in
  i?86) godeb_package="godeb-386.tar.gz" ;;
  x86_64) godeb_package="godeb-amd64.tar.gz" ;;
esac

# Download and unpack "godeb" package.
{
  curl -s https://godeb.s3.amazonaws.com/${godeb_package} | tar zxvf - -C .
} &> /dev/null

# Run "godeb" and install Go Language interpreter.
{
  sudo ./godeb install
} &> /dev/null

echo 'All done!'
}

# Select appropriate Vagrant box per underlying architecture.
box = "precise#{platform_bits}"

# Virtual Machine name et al.
name = "go-magic-#{box}"

Vagrant.configure("2") do |config|
  config.ssh.forward_agent = true
  config.vm.define name.to_sym do |machine|
    machine.vm.box = "hashicorp/#{box}"
    machine.vm.box_check_update = false
    machine.vm.hostname = name
    machine.vm.provider :virtualbox do |vb|
      vb.name = name
      vb.gui = false
      vb.customize ['modifyvm', :id,
        '--memory', '384',
        '--cpus', '1',
        '--rtcuseutc', 'on',
        '--natdnshostresolver1', 'on',
        '--natdnsproxy1', 'on'
      ]
    end
    machine.vm.provision :shell, privileged: false, inline: script
  end
end

# vim: set ts=2 sw=2 sts=2 et :
# encoding: utf-8
