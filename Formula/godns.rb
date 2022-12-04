class Godns < Formula
  desc "DNS server"
  homepage "https://github.com/dhcmrlchtdj/godns"
  license "AGPL-3.0-or-later"
  head "https://github.com/dhcmrlchtdj/godns.git", branch: "main"

  depends_on "rust" => :build

  def install
    system "cargo", "install", *std_cargo_args
    (etc/"godns").mkpath
    etc.install "aur/config.json" => "godns/config.json"
  end

  service do
    run [opt_bin/"godns", "--conf", etc/"godns/config.json"]
    keep_alive true
    log_path var/"log/godns.log"
    error_log_path var/"log/godns.log"
  end

  test do
    system "true"
  end
end
