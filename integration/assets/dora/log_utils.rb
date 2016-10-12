require "logging_service"

class LogUtils < Sinatra::Base

  STDOUT.sync = true

  $run = false
  $sequence_number = 0
  $logging_service = ::LoggingService.new

  get "/loglines/:linecount" do
    produce_log_output(params[:linecount])
    "logged #{params[:linecount]} line to stdout"
  end

  get "/loglines/:linecount/:tag" do
    produce_log_output(params[:linecount], params[:tag])
    "logged #{params[:linecount]} line with tag #{params[:tag]} to stdout"
  end

  get '/log/sleep/count' do
    $logging_service.log_message_count
  end

  get '/log/sleep/running' do
    $logging_service.running
  end

  get '/log/sleep/:logspeed/limit/:limit' do
    limit = params[:limit].to_i
    logspeed = params[:logspeed]
    $logging_service.produce_logspeed_output(limit, logspeed, request.host)
  end

  get '/log/sleep/:logspeed' do
    logspeed = params[:logspeed]
    $logging_service.produce_logspeed_output(0, logspeed, request.host)
  end

  get '/log/bytesize/:bytesize' do
    $run = true
    logString = "0" * params[:bytesize].to_i
    STDOUT.puts("Muahaha... let's go. No wait. Logging #{params[:bytesize]} bytes per logline.")
    while $run do
      STDOUT.puts(logString)
    end
  end

  get '/log/stop' do
    $logging_service.stop
  end

  private
  def produce_log_output(linecount, tag="")
    linecount.to_i.times do |i|
      STDOUT.puts "#{Time.now.strftime("%FT%T.%N%:z")} line #{i} #{tag}"
      $stdout.flush
    end
  end
end
