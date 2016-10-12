require "spec_helper"

describe Curl do
  describe "GET /curl/1.2.3.4" do

    it "should run curl with the host and port" do
      get "/curl/example.com"

      expect(last_response.status).to eq(200)
      expect(last_response.body).to_not be_nil

      response = JSON.parse!(last_response.body)
      ["stdout", "stderr", "return_code"].each do |k|
        expect(response.key?(k)).to be_true
      end

      expect(response["stdout"]).to match(/<html>/i)
      expect(response["return_code"]).to eq(0)
    end

    it "should return error status for failure to resolve host" do
      get "/curl/example.example"

      expect(last_response.status).to eq(200)

      response = JSON.parse!(last_response.body)

      expect(response["stdout"]).to eq("")
      expect(response["return_code"]).to eq(6) # Couldn't resolve host. The given remote host was not resolved.
    end

    it "should return error status for connection failure" do
      get "/curl/127.0.0.1/9999"

      expect(last_response.status).to eq(200)

      response = JSON.parse!(last_response.body)

      expect(response["stdout"]).to eq("")
      expect(response["return_code"]).to eq(7) # Failed to connect to host.
    end
  end
end
