#!/usr/bin/env ruby

require 'json'
require 'pp'
require 'securerandom'

broker_name = ARGV[0]
broker_name ||= 'async-broker'

env = ARGV[1]
env ||= 'bosh-lite'

env_to_domain_mapping = {
  'bosh-lite' => 'bosh-lite.com',
  'a1' => 'a1-app.cf-app.com',
  'tabasco' => 'tabasco-app.cf-app.com'
}

domain = env_to_domain_mapping[env] || env

puts "Setting up broker `#{broker_name}` on #{domain}"

$service_name = nil

def uniquify_config
  puts 'Creating a unique configuration for broker'

  raw_config = File.read('data.json')
  config = JSON.parse(raw_config)
  catalog = config['behaviors']['catalog']['body']

  plan_mapping = {}
  catalog['services'] = catalog['services'].map do |service|
    $service_name = service['name'] = "fake-service-#{SecureRandom.uuid}"
    service['id'] = SecureRandom.uuid

    service['dashboard_client']['id'] = SecureRandom.uuid
    service['dashboard_client']['secret'] = SecureRandom.uuid

    service['plans'] = service['plans'].map do |plan|
      original_id = plan['id']
      plan['id'] = SecureRandom.uuid
      plan_mapping[original_id] = plan['id']
      plan
    end
    service
  end

  config['behaviors'].each do |action, behavior|
    next if action == 'catalog'

    behavior.keys.each do |plan_id|
      next if plan_id == 'default'

      response = behavior[plan_id]
      new_plan_id = plan_mapping[plan_id]
      behavior[new_plan_id] = response
      behavior.delete(plan_id)
    end
  end

  File.open('data.json', 'w') do |file|
    file.write(JSON.pretty_generate(config))
  end
end

def push_broker(broker_name, domain)
  puts "Pushing the broker"
  IO.popen("cf push #{broker_name} -d #{domain}") do |cmd_output|
    cmd_output.each { |line| puts line }
  end
  puts
  puts
end

def create_service_broker(broker_name, url)
  output = []
  IO.popen("cf create-service-broker #{broker_name} user password #{url}") do |cmd|
    cmd.each do |line|
      puts line
      output << line
    end
  end
  output
end

def broker_already_exists?(output)
  output.any? { |line| line =~ /service broker url is taken/ }
end

def update_service_broker(broker_name, url)
  puts
  puts "Broker already exists. Updating"
  IO.popen("cf update-service-broker #{broker_name} user password #{url}") do |cmd|
    cmd.each { |line| puts line }
  end
  puts
end

def enable_service_access
  IO.popen("cf enable-service-access #{$service_name}") do |cmd|
    cmd.each { |line| puts line }
  end
end

uniquify_config
push_broker(broker_name, domain)

url = "http://#{broker_name}.#{domain}"

output = create_service_broker(broker_name, url)
if broker_already_exists?(output)
  update_service_broker(broker_name, url)
end

enable_service_access

puts
puts 'Setup complete'