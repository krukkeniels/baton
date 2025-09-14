class Baton < Formula
  desc "CLI Orchestrator for LLM-Driven Task Execution"
  homepage "https://github.com/krukkeniels/baton"
  url "https://github.com/krukkeniels/baton/releases/latest/download/baton-darwin-amd64"
  version "1.0.0"
  sha256 "YOUR_SHA256_HERE"  # Will be updated by release automation

  on_arm do
    url "https://github.com/krukkeniels/baton/releases/latest/download/baton-darwin-arm64"
    sha256 "YOUR_ARM64_SHA256_HERE"
  end

  def install
    bin.install "baton-darwin-amd64" => "baton" if Hardware::CPU.intel?
    bin.install "baton-darwin-arm64" => "baton" if Hardware::CPU.arm?
  end

  test do
    system "#{bin}/baton", "--version"
  end

  def caveats
    <<~EOS
      Baton requires Claude Code CLI for full functionality:
        Visit https://claude.ai/code for installation instructions

      Get started with:
        baton init
        baton --help

      Documentation: https://github.com/krukkeniels/baton#readme
    EOS
  end
end