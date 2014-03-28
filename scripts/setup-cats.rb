#!/usr/bin/env ruby

require 'open3'

BOSH_LITE_HOSTNAME = ENV['BOSH_LITE_HOSTNAME']
CC_HOSTNAME = ENV['CC_HOSTNAME']
CF_ADMIN_USER = ENV['CF_ADMIN_USER']
CF_ADMIN_PASSWORD = ENV['CF_ADMIN_PASSWORD']

def cf(cmd)
  run_or_die("out/cf #{cmd}")
end

def run_or_die(cmd)
  puts "Running #{cmd}"
  out, err, status = Open3.capture3(cmd)
  puts out
  raise "Command failed:\n#{err}" unless status == 0
end

cleanup_cmd = <<-BASH
sudo rm -rfv $(sudo find /opt/warden/disks/bind_mount_points -name "*cc-droplets*") && \
sudo truncate -s 0 /opt/warden/disks/ephemeral_mount_point/*/sys/log/dea_logging_agent/*.log
BASH
run_or_die(%(ssh -o "StrictHostKeyChecking no" #{BOSH_LITE_HOSTNAME} '#{cleanup_cmd}'))

cf "api #{CC_HOSTNAME} --skip-ssl-validation"
cf "login -u #{CF_ADMIN_USER} -p #{CF_ADMIN_PASSWORD}"

%w(linux32 linux64 win32 win64 osx).each do |platform|
  user = "cats-user-#{platform}"
  org = "cats-org-#{platform}"
  cf "create-user #{user} cats-password" rescue puts "create-user failed, but still continuing"
  cf "delete-org -f #{org}"
  cf "create-org #{org}"

  ["cats-space-#{platform}", "persistent-space"].each do |space|
    cf "create-space #{space} -o #{org}"

    %w(SpaceManager SpaceDeveloper SpaceAuditor).each do |role|
      cf "set-space-role #{user} #{org} #{space} #{role}"
    end
  end
end
