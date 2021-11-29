class Godns < Formula
  desc "DNS with china list"
  homepage "https://github.com/dhcmrlchtdj/godns"
  license "AGPL-3.0-or-later"
  head "https://github.com/dhcmrlchtdj/godns.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args
    (etc/"godns").mkpath
    etc.install "aur/config.json.example" => "godns/config.json.example"
  end

  service do
    run [opt_bin/"godns", "-conf", etc/"godns/config.json"]
    keep_alive true
    log_path var/"log/godns.log"
    error_log_path var/"log/godns.log"
  end

  test do
    system "true"
  end
end
