ENV['RACK_ENV'] ||= 'development'

require 'rubygems'
require 'sinatra/base'
require 'json'
require 'pp'
require 'logger'
require 'rainbow/ext/string'

require 'bundler'
Bundler.require :default, ENV['RACK_ENV'].to_sym
Rainbow.enabled = true

$stdout.sync = true
$stderr.sync = true

class ServiceInstance
  attr_reader :provision_data, :fetch_count, :deleted

  def initialize(opts={})
    @provision_data = opts.fetch(:provision_data)
    @fetch_count = opts.fetch(:fetch_count, 0)
    @deleted = opts.fetch(:deleted, false)
  end

  def plan_id
    @provision_data['plan_id']
  end

  def update!(updated_data)
    @provision_data.merge!(updated_data)
    @fetch_count = 0
    self
  end

  def delete!
    @deleted = true
    @fetch_count = 0
    self
  end

  def increment_fetch_count
    @fetch_count += 1
  end

  def to_json(opts={})
    {
      provision_data: provision_data,
      fetch_count: fetch_count,
      deleted: deleted
    }.to_json(opts)
  end
end

class DataSource
  attr_reader :data

  def initialize(data = nil)
    @data = data || JSON.parse(File.read(File.absolute_path('data.json')))
  end

  def max_fetch_service_instance_requests
    @data['max_fetch_service_instance_requests'] || 1
  end

  def service_instance_by_id(cc_id)
    @data['service_instances'][cc_id]
  end

  def create_service_instance(cc_id, json_data)
    service_instance = ServiceInstance.new(
      provision_data: json_data,
    )

    @data['service_instances'][cc_id] = service_instance

    service_instance
  end

  def create_service_binding(instance_id, binding_id, binding_data)
    @data['service_instances'][binding_id] = {
      'binding_data' => binding_data,
      'instance_id' => instance_id,
    }
  end

  def delete_service_binding(binding_id)
    @data['service_instances'].delete(binding_id)
  end

  def merge!(data)
    data = data.dup
    data['service_instances'] = data.fetch('service_instances', {}).inject({}) do |service_instances, (guid, instance_data)|
      symbolized_data = instance_data.inject({}) do |memo,(k,v)|
        memo[k.to_sym] = v
        memo
      end

      service_instances[guid] = ServiceInstance.new(symbolized_data)
      service_instances
    end

    data.each_pair do |key, value|
      if @data[key] && @data[key].is_a?(Hash)
        @data[key].merge!(value)
      else
        @data[key] = value
      end
    end
  end

  def without_instances_or_bindings
    @data.reject { |key| %w(service_instances service_bindings).include?(key) }
  end

  def behavior_for_type(type, plan_id)
    plans_or_default_behavior = @data['behaviors'][type.to_s]

    return plans_or_default_behavior if type == :catalog

    raise "Behavior object is missing key: #{type} (tried to lookup plan_id #{plan_id})" unless plans_or_default_behavior

    if plan_id && plans_or_default_behavior.has_key?(plan_id)
      plans_or_default_behavior[plan_id]
    else
      $log.info("Could not find response for plan id: #{plan_id}")
      return plans_or_default_behavior['default'] if plans_or_default_behavior['default']
      raise "Behavior for #{type} is missing response for plan_id #{plan_id} and default response."
    end
  end
end

class ServiceBroker < Sinatra::Base
  set :logging, true

  configure :production, :development, :test do
    $datasource = DataSource.new
    $log = Logger.new(STDOUT)
    $log.level = Logger::INFO
    $log.formatter = proc do |severity, datetime, progname, msg|
      "#{severity}: #{msg}\n"
    end
  end

  def log(request)
    $log.info "#{request.env['REQUEST_METHOD']} #{request.env['PATH_INFO']} #{request.env['QUERY_STRING']}".color(:yellow)
    request.body.rewind
    headers = find_headers(request)
    $log.info "Request headers: #{headers}".color(:cyan)
    $log.info "Request body: #{request.body.read}".color(:yellow)
    request.body.rewind
  end

  def find_headers(request)
    request.env.select { |key, _| key =~ /HTTP/ }
  end

  def log_response(status, body)
    $log.info "Response: status=#{status}, body=#{body}".color(:green)
    body
  end

  def respond_with_behavior(behavior, accepts_incomplete=false)
    sleep behavior['sleep_seconds']

    if behavior['async_only'] && !accepts_incomplete
      respond_async_required
    else
      respond_from_config(behavior)
    end
  end

  def respond_async_required
    status 422
    log_response(status, {
      'error' => 'AsyncRequired',
      'description' => 'This service plan requires client support for asynchronous service operations.'
     }.to_json)
  end

  def respond_from_config(behavior)
    status behavior['status']
    if behavior['body']
      log_response(status, behavior['body'].to_json)
    else
      log_response(status, behavior['raw_body'])
    end
  end

  before do
    log(request)
  end

  # fetch catalog
  get '/v2/catalog/?' do
    respond_with_behavior($datasource.behavior_for_type(:catalog, nil))
  end

  # provision
  put '/v2/service_instances/:id/?' do |id|
    json_body = JSON.parse(request.body.read)
    service_instance = $datasource.create_service_instance(id, json_body)
    respond_with_behavior($datasource.behavior_for_type(:provision, service_instance.plan_id), params['accepts_incomplete'])
  end

  # fetch service instance
  get '/v2/service_instances/:id/last_operation/?' do |id|
    service_instance = $datasource.service_instance_by_id(id)
    if service_instance
      plan_id = service_instance.plan_id

      if service_instance.increment_fetch_count > $datasource.max_fetch_service_instance_requests
        state = 'finished'
      else
        state = 'in_progress'
      end

      behavior = $datasource.behavior_for_type('fetch', plan_id)[state]
      sleep behavior['sleep_seconds']
      status behavior['status']

      if behavior['body']
        log_response(status, behavior['body'].to_json)
      else
        log_response(status, behavior['raw_body'])
      end
    else
      status 200
      log_response(status, {
        state: 'failed',
        description: "Broker could not find service instance by the given id #{id}",
      }.to_json)
    end
  end

  # fetch service binding
  get '/v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation/?' do |instance_id, binding_id|
    status 200
    log_response(status, {
      state: 'succeeded',
      description: '100%',
    }.to_json)
  end

  # update service instance
  patch '/v2/service_instances/:id/?' do |id|
    json_body = JSON.parse(request.body.read)
    service_instance = $datasource.service_instance_by_id(id)
    plan_id = json_body['plan_id']

    behavior = $datasource.behavior_for_type(:update, plan_id)
    if [200, 202].include?(behavior['status'])
      service_instance.update!(json_body) if service_instance
    end

    respond_with_behavior(behavior, params['accepts_incomplete'])
  end

  # deprovision
  delete '/v2/service_instances/:id/?' do |id|
    service_instance = $datasource.service_instance_by_id(id)
    if service_instance
      service_instance.delete!
      respond_with_behavior($datasource.behavior_for_type(:deprovision, service_instance.plan_id), params[:accepts_incomplete])
    else
      respond_with_behavior($datasource.behavior_for_type(:deprovision, nil), params[:accepts_incomplete])
    end
  end

  # create service binding
  put '/v2/service_instances/:instance_id/service_bindings/:id' do |instance_id, binding_id|
    content_type :json
    json_body = JSON.parse(request.body.read)

    service_binding = $datasource.create_service_binding(instance_id, binding_id, json_body)
    respond_with_behavior($datasource.behavior_for_type(:bind, service_binding['binding_data']['plan_id']), params[:accepts_incomplete])
  end

  # delete service binding
  delete '/v2/service_instances/:instance_id/service_bindings/:id' do |instance_id, binding_id|
    content_type :json

    service_binding = $datasource.delete_service_binding(binding_id)
    if service_binding
      respond_with_behavior($datasource.behavior_for_type(:unbind, service_binding['binding_data']['plan_id']), params[:accepts_incomplete])
    else
      respond_with_behavior($datasource.behavior_for_type(:unbind, nil), params[:accepts_incomplete])
    end
  end

  get '/v2/service_instances/:instance_id' do |instance_id|
    status 200
    log_response(status, JSON.pretty_generate($datasource.data['service_instances'][instance_id].provision_data))
  end

  get '/v2/service_instances/:instance_id/service_bindings/:id' do |instance_id, binding_id|
    binding_data = $datasource.data['service_instances'][binding_id]['binding_data']
    response_body = $datasource.behavior_for_type(:fetch_service_binding, binding_data['plan_id'])
    response_body['body'].merge!(binding_data)
    respond_with_behavior(response_body)
  end

  get '/config/all/?' do
    log_response(status, JSON.pretty_generate($datasource.data))
  end

  get '/config/?' do
    log_response(status, JSON.pretty_generate($datasource.without_instances_or_bindings))
  end

  post '/config/?' do
    json_body = JSON.parse(request.body.read)
    $datasource.merge!(json_body)
    log_response(status, JSON.pretty_generate($datasource.without_instances_or_bindings))
  end

  post '/config/reset/?' do
    $datasource = DataSource.new
    log_response(status, JSON.pretty_generate($datasource.without_instances_or_bindings))
  end

  error do
    status 500
    e = env['sinatra.error']
    log_response(status, JSON.pretty_generate({
      error: true,
      message: e.message,
      path: request.url,
      timestamp: Time.new,
      type: '500',
      backtrace: e.backtrace
    }))
  end

  run! if app_file == $0
end
