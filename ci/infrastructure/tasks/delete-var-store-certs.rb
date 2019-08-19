#!/usr/bin/env ruby

require 'yaml'
require 'pp'

yaml_hash = YAML.load_file(ARGV[0])
yaml_hash.delete_if{|key, value| value.is_a?(Hash) && value.key?('certificate')}

File.open(ARGV[0], 'w+') do |file|
  file.write(YAML.dump(yaml_hash))
end
