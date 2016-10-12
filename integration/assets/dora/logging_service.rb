class LoggingService

  def initialize
    @run = false
    @sequence_number = 0
    @output = STDOUT
  end

  def output= (output_override)
    @output = output_override
  end

  def running
    @run.to_s
  end

  def log_message_count
    @sequence_number.to_s
  end

  def stop
    @output.puts("Stopped logs #{Time.now}")
    @run = false
  end

  def produce_logspeed_output(limit, logspeed, host_string)
    @run = true
    @sequence_number = 1
    @output.puts("Muahaha... let's go. Waiting #{logspeed.to_f/1000000.to_f} seconds between loglines. Logging 'Muahaha...' every time.")
    while @run do
      sleep(logspeed.to_f/1000000.to_f)
      @output.puts("Log: #{host_string} Muahaha...#{@sequence_number}...#{Time.now}")
      break if (limit > 0) && (@sequence_number >= limit)
      @sequence_number += 1
    end
    @run = false
  end
end