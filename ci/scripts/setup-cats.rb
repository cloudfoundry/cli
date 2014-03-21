#!/usr/bin/env ruby

require 'open3'

BOSH_LITE_HOSTNAME = ENV['BOSH_LITE_HOSTNAME']
CC_HOSTNAME = ENV['CC_HOSTNAME']
CF_ADMIN_USER = ENV['CF_ADMIN_USER']
CF_ADMIN_PASSWORD = ENV['CF_ADMIN_PASSWORD']

def cf(cmd)
  out, err, status = Open3.capture3("out/cf #{cmd}")
  puts out
  raise "cf failed:\n#{err}" unless status == 0
end

system "ssh #{BOSH_LITE_HOSTNAME} \"rm -rf $(find /opt/warden/disks/bind_mount_points -name '*cc-droplets*' 2> /dev/null)\""

cf "api #{CC_HOSTNAME} --skip-ssl-validation"
cf "auth #{CF_ADMIN_USER} #{CF_ADMIN_PASSWORD}"

%w(linux32 linux64 win32 win64 osx).each do |platform|
  user = "cats-user-#{platform}"
  org = "cats-org-#{platform}"
  cf "create-user #{user} cats-password"
  cf "delete-org -f #{org}"
  cf "create-org #{org}"

  ["cats-space-#{platform}", "persistent-space-#{platform}"].each do |space|
    cf "create-space #{space} -o #{org}"

    %w(SpaceManager SpaceDeveloper SpaceAuditor).each do |role|
      cf "set-space-role #{user} #{org} #{space} #{role}"
    end
  end
end
