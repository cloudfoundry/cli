#!/usr/bin/env ruby

require 'CSV'
require 'json'
require 'benchmark'
require 'securerandom'
require 'optparse'

DEFAULT_BROKER_URL = 'http://async-broker.bosh-lite.com'

def get_config
  raw_config = File.read('data.json')
  JSON.parse(raw_config)
end

def get_service
  config = get_config
  config['behaviors']['catalog']['body']['services'].first['name']
end

def get_plan
  config = get_config
  config['behaviors']['catalog']['body']['services'].first['plans'].first['name']
end

def get_second_plan
  config = get_config
  config['behaviors']['catalog']['body']['services'].first['plans'][1]['name']
end

def execute(cmd)
  `#{cmd}`
end

class ProvisionCommand
  def setup(instance_name)
  end

  def run(instance_name)
    execute "cf create-service #{get_service} #{get_plan} #{instance_name}"
  end

  def cleanup(instance_name)
  end
end

class UpdateCommand
  def setup(instance_name)
    execute "cf create-service #{get_service} #{get_plan} #{instance_name}"
  end

  def run(instance_name)
    execute "cf update-service #{instance_name} -p #{get_second_plan}"
  end

  def cleanup(instance_name)
  end
end

class DeprovisionCommand
  def setup(instance_name)
    execute "cf create-service #{get_service} #{get_plan} #{instance_name}"
  end

  def run(instance_name)
    execute "cf delete-service #{instance_name} -f"
  end

  def cleanup(instance_name)
  end
end

class CleanupCommandWrapper
  def initialize(command, broker_url)
    @command = command
    @broker_url = broker_url
  end

  def setup(instance_name)
    @command.setup(instance_name)
  end

  def run(instance_name)
    @command.run(instance_name)
  end

  def cleanup(instance_name)
    @command.cleanup(instance_name)
    -> {
      execute "curl -s #{@broker_url}/config/reset -X POST"
      until attempt_delete(instance_name)
      end
    }
  end

  private

  def attempt_delete(instance_name)
    output = execute "cf delete-service #{instance_name} -f"
    !output.include?('Another operation for this service instance is in progress')
  end
end

def write_output_file(output_file, rows)
  CSV.open(output_file, 'w') do |csv|
    csv << rows[0].headers
    rows.each do |row|
      csv << row
    end
  end
end

def delete_leftover_instances(deferred_deletions)
  count = deferred_deletions.compact.count
  STDOUT.write("Cleaning up service instances ... 0 / #{count}")
  STDOUT.flush
  i = 0
  deferred_deletions.compact.each do |callback|
    callback.call
    i += 1
    STDOUT.write("\rCleaning up service instances ... #{i} / #{count}")
    STDOUT.flush
  end
  puts
  puts "Done"
end

def parse_parameters
  options = { cleanup: true }
  OptionParser.new do |opts|
    opts.on("--no-cleanup", "Run script without cleanup") do |v|
      options[:cleanup] = v
    end
  end.parse!

  if ARGV.length < 1
    puts "Usage: #{$PROGRAM_NAME} CSV_FILE [BROKER_URL] [--no-cleanup]"
    puts
    puts "Broker URL defaults to #{DEFAULT_BROKER_URL}"
    exit(1)
  end

  input_file = ARGV[0]

  name = File.basename(input_file, '.*')
  extension = File.extname(input_file)
  output_file = name + "-out" + extension

  broker_url = ARGV.length > 1 ? ARGV[1] : DEFAULT_BROKER_URL
  return broker_url, input_file, output_file, options[:cleanup]
end

def configure_broker_endpoint(action, body, broker_url, row, status)
  json_config = {
    behaviors: {
      action => {
        default: {
          status: status,
          raw_body: body,
          sleep_seconds: row['sleep seconds'].to_f
        }
      }
    }
  }

  execute "curl -s #{broker_url}/config/reset -X POST"
  execute "curl -s #{broker_url}/config -d '#{json_config.to_json}'"
end

def run_command(command, deferred_deletions, cleanup, line_number)
  instance_name = "si-#{line_number}-#{SecureRandom.uuid}"

  command.setup(instance_name)
  output = command.run(instance_name)
  deferred_deletions << command.cleanup(instance_name) if cleanup
  output
end

deferred_deletions = []
rows = []

broker_url, input_file, output_file, cleanup = parse_parameters

action_to_cmd_mapping = {
  provision: CleanupCommandWrapper.new(ProvisionCommand.new, broker_url),
  update: CleanupCommandWrapper.new(UpdateCommand.new, broker_url),
  deprovision: CleanupCommandWrapper.new(DeprovisionCommand.new, broker_url),
}

report = Benchmark.measure do
  i = 0
  CSV.foreach(input_file, headers: true) do |row|
    i += 1
    rows << row

    action, status, body = row['action'], row['status'], row['body']
    next unless action

    command = action_to_cmd_mapping[action.to_sym]
    next unless command

    configure_broker_endpoint(action, body, broker_url, row, status)

    output = run_command(command, deferred_deletions, cleanup, i)
    row['output'] = output
    STDOUT.write('.')
    STDOUT.flush
  end

  puts

  write_output_file(output_file, rows)
  delete_leftover_instances(deferred_deletions)
end

puts "Took #{report.real} seconds"
