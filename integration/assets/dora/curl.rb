class Curl < Sinatra::Base

  get '/curl/:host/?:port?' do
    host = params[:host]
    port = params[:port] || "80"

    stdout, stderr, status = Open3.capture3("curl -m 3 -v -i #{host}:#{port}")

    { stdout: stdout, stderr: stderr, return_code: status.exitstatus }.to_json
  end

end
