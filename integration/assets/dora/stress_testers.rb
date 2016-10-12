require "open3"

class StressTesters < Sinatra::Base
  ACCEPTED_OPTIONS = %w[timeout cpu io vm vm-bytes vm-stride vm-hang vm-keep hdd hdd-bytes].freeze
  STANDARD_OPTIONS = %w[instance_id splat captures].freeze

  helpers do
    def run(command)
      output = []
      exit_status = 0

      Open3.popen2e(command) do |_, stdout_and_stderr, wait_thr|
        output += stdout_and_stderr.readlines
        exit_status = wait_thr.value
      end

      response_status = (exit_status == 0 ? 200 : 500)
      [response_status, output.join]
    end
  end

  get "/stress_testers" do
    run('pgrep stress | xargs -r ps -H')
  end

  post "/stress_testers" do
    command = ["./stress"]

    params.each do |option, value|
      next if STANDARD_OPTIONS.include?(option)
      halt 412 unless ACCEPTED_OPTIONS.include?(option)
      command << "--#{option} #{value}"
    end

    pid = Process.spawn(command.join(" "), in: "/dev/null", out: "/dev/null", err: "/dev/null")
    Process.detach(pid)
    [201]
  end

  delete "/stress_testers" do
    run('pkill stress')
  end
end
