require "spec_helper"

describe LoggingService do

  let(:fake_io) { StringIO.new }

  let(:logging_service) {
    logging_service = LoggingService.new
    logging_service.output = fake_io
    logging_service
  }

  describe "produce_logspeed_output" do
    it "should write log message with the log speed to the output" do
      logging_service.produce_logspeed_output(1, 1000, "foo")
      expect(fake_io.string).to include("Waiting 0.001 seconds between loglines.")
    end

    it "should write log message with the host to the output" do
      logging_service.produce_logspeed_output(1, 1000, "dora app 1")
      expect(fake_io.string).to include("Log: dora app 1")
    end

    it "logs until stop gets called if limit is 0" do
      expect(logging_service.running).to eq("false")
      Thread.new do
        logging_service.produce_logspeed_output(0, 1000, "dora app 1")
      end

      eventually { expect(logging_service.running).to eq("true") }
      logging_service.stop
      expect(logging_service.running).to eq("false")
    end
  end

  describe "stop"
end