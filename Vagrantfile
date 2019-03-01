# -*- mode: ruby -*-
# vi: set ft=ruby :

# Prerequisites validation

## Vagrant version
Vagrant.require_version ">= 1.7.4"

$forwarded_ports = { }

# Make sure the vagrant-ignition plugin is installed
required_plugins = %w{ }

plugins_to_install = required_plugins.select { |plugin| not Vagrant.has_plugin? plugin }
if not plugins_to_install.empty?
  puts "Installing plugins: #{plugins_to_install.join(' ')}"
  if system "vagrant plugin install #{plugins_to_install.join(' ')}"
    exec "vagrant #{ARGV.join(' ')}"
  else
    abort "Installation of one or more plugins has failed. Aborting."
  end
end

set_environment_variables = <<SCRIPT
  cat << EOF > /etc/profile.d/myvars.sh
# environment variables.
# change default docker-compose load file name
export COMPOSE_FILE=docker-compose.yml
alias dc='docker-compose'
EOF
SCRIPT

latest_docker_install_script = <<SCRIPT
  DOCKER_VERSION=18.09.2
  DOCKER_COMPOSE_VERSION=1.23.2

  docker version
  /etc/init.d/docker restart $DOCKER_VERSION
  wget -q -L https://github.com/docker/compose/releases/download/$DOCKER_COMPOSE_VERSION/docker-compose-`uname -s`-`uname -m`
  mv docker-compose-`uname -s`-`uname -m` /opt/bin/docker-compose
  chmod +x /opt/bin/docker-compose
  chown root:root /opt/bin/docker-compose
SCRIPT

fix_dns_use_ipv6 = <<SCRIPT
  sed -i "s/^nameserver 8.8.8.8$/#nameserver 8.8.8.8/g" /etc/resolv.conf
  sed -i "s/^nameserver 8.8.4.4$/#nameserver 8.8.4.4/g" /etc/resolv.conf
SCRIPT

ssh_path_init = <<SCRIPT
  cat << EOF > /home/bargee/.bash_profile
if [ -f "/home/bargee/.bashrc" ]; then
  source "/home/bargee/.bashrc"
fi
cd /vagrant
EOF
SCRIPT

module VagrantPlugins
  module GuestLinux
    class Plugin < Vagrant.plugin("2")
      guest_capability("linux", "change_host_name") { Cap::ChangeHostName }
      guest_capability("linux", "configure_networks") { Cap::ConfigureNetworks }
    end
  end
end

Vagrant.configure("2") do |config|
  config.vm.define "barge"
  config.vm.box = "ailispaw/barge"

  $forwarded_ports.each do |guest, host|
    config.vm.network "forwarded_port", guest: guest, host: host, auto_correct: true
  end

  config.vm.provider "virtualbox" do |v|
    v.memory = 1024
    v.cpus = 2
  end

  config.vm.synced_folder ".", "/vagrant"

# for NFS synced folder
#  config.vm.network "private_network", ip: "192.168.33.10"
#  config.vm.synced_folder ".", "/vagrant", type: "nfs"
#    mount_options: ["nolock", "vers=3", "udp", "noatime", "actimeo=1"]

#  for RSync synced folder
#  config.vm.synced_folder ".", "/vagrant", type: "rsync",
#    rsync__args: ["--verbose", "--archive", "--delete", "--copy-links"]

  config.vm.provision :shell, :inline => latest_docker_install_script
  config.vm.provision :shell, :inline => ssh_path_init
  config.vm.provision :shell, :inline => fix_dns_use_ipv6, run: "always"
  config.vm.provision :shell, :inline => set_environment_variables, run: "always"
end
