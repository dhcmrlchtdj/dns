class Godns < Formula
  desc "DNS with china list"
  homepage "https://github.com/dhcmrlchtdj/godns"
  license "AGPL-3.0-or-later"
  head "https://github.com/dhcmrlchtdj/godns.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args
    etc.install "aur/config.json.example" => "godns/config.json.example"
  end

  plist_options manual: "#{HOMEBREW_PREFIX}/opt/dns/bin/godns"

  def plist
    <<~EOS
      <?xml version="1.0" encoding="UTF-8"?>
      <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
      <plist version="1.0">
        <dict>
          <key>Label</key>
          <string>#{plist_name}</string>
          <key>ProgramArguments</key>
            <array>
                <string>#{opt_bin}/godns</string>
                <string>-conf</string>
                <string>#{etc}/godns/config.json</string>
            </array>
            <key>RunAtLoad</key>
            <true/>
            <key>KeepAlive</key>
            <true/>
            <key>StandardOutPath</key>
            <string>#{var}/log/godns.log</string>
            <key>StandardErrorPath</key>
            <string>#{var}/log/godns.log</string>
          </dict>
      </plist>
    EOS
  end

  test do
    system "true"
  end
end
