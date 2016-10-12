require "spec_helper"

describe StressTesters do
  let(:pid) { 23 }
  let(:wait_thread) { double(:fake_wait_thread, value: 0) }

  before do
    @spawn_commands = []
    @open3_commands = []
    allow(Process).to receive(:spawn) do |command, options|
      @spawn_commands << {command: command, options: options}
      pid
    end
    allow(Process).to receive(:detach)

    allow(Open3).to receive(:popen2e) do |command, &block|
      @open3_commands << command
      output = @spawn_commands.map { |command| command[:command] }.join("\n")
      block.call nil, StringIO.new(output), wait_thread
    end
  end

  describe "GET/POST /stress_testers" do
    it "should detach the process" do
      expect(Process).to receive(:detach).with(pid)
      post "/stress_testers"
    end

    it "creates a new stress process with default params" do
      get "/stress_testers"
      expect(last_response.status).to eq 200
      expect(last_response.body).to eq ""
      expect(@open3_commands).to include("pgrep stress | xargs -r ps -H")

      post "/stress_testers"
      expect(last_response.status).to eq 201
      expect(@spawn_commands).to include(command: "./stress", options: {in: "/dev/null", out: "/dev/null", err: "/dev/null"})

      get "/stress_testers"
      expect(last_response.status).to eq 200
      expect(last_response.body).to include "stress"
    end

    context "when trying to customize the load" do
      %w[timeout cpu io vm vm-bytes vm-stride vm-hang vm-keep hdd hdd-bytes].each do |option|
        it "invokes stress with specified #{option} load" do
          post "/stress_testers?#{option}=23"
          expect(last_response.status).to eq 201

          get "/stress_testers"
          expect(last_response.status).to eq 200
          expect(last_response.body).to include "stress --#{option} 23"
        end
      end

      it "when invoking with more than one option" do
        post "/stress_testers?vm=33&cpu=22"
        expect(last_response.status).to eq 201

        get "/stress_testers"
        expect(last_response.status).to eq 200
        expect(last_response.body).to include "stress"
        expect(last_response.body).to include "--vm 33"
        expect(last_response.body).to include "--cpu 22"
      end

      it "copes when an invalid option is requested" do
        post "/stress_testers?bad_option=22"
        expect(last_response.status).to eq 412
      end
    end

    context "when the pgrep command fails" do
      let(:wait_thread) { double(:fake_wait_thread, value: 1) }

      it "copes with invalid commands" do
        get "/stress_testers"
        expect(last_response.status).to eq 500
      end
    end
  end

  describe "DELETE /stress_testers" do
    it "stops a stress process on the instance" do
      delete "/stress_testers"
      expect(last_response.status).to eq 200
      expect(@open3_commands).to include("pkill stress")
    end

    context "when the pkill command fails" do
      let(:wait_thread) { double(:fake_wait_thread, value: 1) }

      it "copes with invalid commands" do
        delete "/stress_testers"
        expect(last_response.status).to eq 500
      end
    end
  end
end