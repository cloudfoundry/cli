ENV['RACK_ENV'] = 'test'

$: << File.expand_path("../../.", __FILE__)

require 'service_broker'
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
  #conf.include AsyncHelper

  def app
    ServiceBroker
  end
end
