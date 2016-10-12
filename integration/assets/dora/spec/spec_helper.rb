ENV['RACK_ENV'] = 'test'
ENV['VCAP_APPLICATION'] = '{"instance_id":"b4ffb1a7b677447296f9ff7eea535c43","instance_index":0,"host":"0.0.0.0","port":61167,"started_at":"2013-12-06 23:10:27 +0000","started_at_timestamp":1386371427,"start":"2013-12-06 23:10:27 +0000","state_timestamp":1386371427,"limits":{"mem":128,"disk":1024,"fds":16384},"application_version":"ac62bd8b-e8ff-4b18-9341-df578d8f7ec0","application_name":"dora","application_uris":["dora.sunset.cf-app.com"],"version":"ac62bd8b-e8ff-4b18-9341-df578d8f7ec0","name":"dora","uris":["dora.sunset.cf-app.com"],"users":null}'

$: << File.expand_path("../../.", __FILE__)

require 'dora'
require 'rspec'
require 'rack/test'

module AsyncHelper
  def eventually(options = {})
    timeout = options[:timeout] || 2
    interval = options[:interval] || 0.1
    time_limit = Time.now + timeout
    loop do
      begin
        yield
      rescue => error
      end
      return if error.nil?
      raise error if Time.now >= time_limit
      sleep interval
    end
  end
end

RSpec.configure do |conf|
  conf.include Rack::Test::Methods
  conf.include AsyncHelper

  def app
    Dora
  end
end

