require "spec_helper"

describe Instances do
  describe "GET /id" do
    let(:id) { "b4ffb1a7b677447296f9ff7eea535c43" }

    it "should get the instance id from the VCAP_APPLICATION json" do
      get "/id"
      expect(last_response.body).to eq ID
      expect(last_response.headers["Set-Cookie"]).to be_nil
    end

    it "should set the JSESSIONID so that we can get a sticky session" do
      post "/session"
      expect(last_response.headers["Set-Cookie"]).to eq "JSESSIONID=#{id}"
      expect(last_response.body).to eq "Please read the README.md for help on how to use sticky sessions."
    end
  end
end