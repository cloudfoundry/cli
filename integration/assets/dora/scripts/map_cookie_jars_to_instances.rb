#!/usr/bin/env ruby
require "json"
DEFAULT_APPLICATION_NAME = "dora"
DEFAULT_APPLICATION_URL = "dora.sunset.cf-app.com"
app_name = ENV['EXPERIMENT_APP_NAME'] || DEFAULT_APPLICATION_NAME
app_url = ENV['EXPERIMENT_APP_URL'] || DEFAULT_APPLICATION_URL

STDERR.puts "Using app '#{app_name}' on #{app_url}."
instances_json = `CF_TRACE=true cf app #{app_name} | grep fds_quota`

instance_stats = JSON.parse(instances_json)

instance_hosts = Hash[instance_stats.map { |instance_number, stats| [instance_number.to_i, { ip_address: stats["stats"]["host"] }] }]

cookie_jars = Dir["cookie_jars/*cjar"]

cookie_jars.each do |jar_filename|
  response_json = `curl -s #{app_url}/env/VCAP_APPLICATION -b #{jar_filename}`
  parsed_response = JSON.parse(response_json)
  instance_index = parsed_response["instance_index"]

  instance_hosts[instance_index.to_i][:cookie_jar] = jar_filename
end

instances = instance_hosts.map do | k, v |
  {index: k, ip_address: v[:ip_address], cookie_jar: v[:cookie_jar]}
end

instances.sort_by! {|inst| inst[:ip_address]}

instances.each do |row|
  puts "#{row[:cookie_jar]} #{row[:index]} #{row[:ip_address]}"
end
