require 'formula'

class CfBin < Formula
  homepage 'https://github.com/cloudfoundry/cli'
  url 'https://github.com/cloudfoundry/cli/releases/download/v6.0.1/cf-darwin-amd64.tgz'
  sha1 '548a83996ade1fb4c4334e4ebcfd558434c01daf'

  def install
    system 'curl -O https://raw.github.com/cloudfoundry/cli/v6.0.1/LICENSE'
    bin.install 'cf-darwin-amd64' => 'cf'
    doc.install 'LICENSE'
  end

  test do
    system "#{bin}/cf"
  end
end
