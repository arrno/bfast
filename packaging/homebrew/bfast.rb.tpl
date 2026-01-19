# Copy this file into blazingly/homebrew-tap when bootstrapping the tap.
# Replace the ALL_CAPS placeholders with the actual tag (vX.Y.Z), version, and checksums.
class Bfast < Formula
  desc "CLI badger for registering repos with blazingly.fast and inserting the badge"
  homepage "https://blazingly.fast"
  version "VERSION"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/arrno/bfast/releases/download/TAG/bfast_VERSION_darwin_arm64.tar.gz"
      sha256 "SHA256_DARWIN_ARM64"
    else
      url "https://github.com/arrno/bfast/releases/download/TAG/bfast_VERSION_darwin_amd64.tar.gz"
      sha256 "SHA256_DARWIN_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/arrno/bfast/releases/download/TAG/bfast_VERSION_linux_arm64.tar.gz"
      sha256 "SHA256_LINUX_ARM64"
    else
      url "https://github.com/arrno/bfast/releases/download/TAG/bfast_VERSION_linux_amd64.tar.gz"
      sha256 "SHA256_LINUX_AMD64"
    end
  end

  def install
    bin.install "bfast"
  end

  test do
    system "#{bin}/bfast", "--help"
  end
end
