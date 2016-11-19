# Provisioning script which installs development dependencies.
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
box = "precise#{lambda { ['x'].pack('P').size * 8 }.()}"

# Virtual Machine name et al.
name = "go-magic-#{box}"

Vagrant.configure("2") do |config|
  config.ssh.forward_agent = true

  if config.vm.respond_to? :box_check_update
    config.vm.box_check_update = true
  end

  if config.vm.respond_to? :use_linked_clone
    config.use_linked_clone = true
  end

  config.vm.define name.to_sym do |machine|
    machine.vm.box = "hashicorp/#{box}"
    machine.vm.hostname = name

    machine.vm.provider :virtualbox do |vb|
      vb.linked_clone = true if Vagrant::VERSION =~ /^1.8/
      vb.name = name
      vb.gui = false
      vb.customize ['modifyvm', :id,
        '--memory', '512',
        '--cpus', '1',
        '--rtcuseutc', 'on',
        '--natdnshostresolver1', 'on',
        '--natdnsproxy1', 'on',
        '--nictype1', 'virtio'
      ]
    end

    machine.vm.provision :shell, privileged: false, inline: script
  end
end
