require 'formula'

class CfBin < Formula
  homepage 'https://github.com/cloudfoundry/cli'
  url 'https://github.com/cloudfoundry/cli/releases/download/v6.0.0/cf-darwin-amd64.tgz'
  sha1 'f2b27f521da2abeacc289071f5f409cfef8ce9e3'

  def install
    system 'curl -O https://raw.github.com/cloudfoundry/cli/v6.0.0/LICENSE'
    bin.install 'cf-darwin-amd64' => 'cf'
    doc.install 'LICENSE'
  end

  test do
    system "#{bin}/cf"
  end
end
