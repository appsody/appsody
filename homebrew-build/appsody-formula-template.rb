class Appsody < Formula
  desc "The Appsody command-line interface"
  homepage "https://www.appsody.dev"
  url "https://github.com/REPO_NAME/releases/download/RELEASE_TAG/FILE_PREFIX-VERSION_NUMBER.tar.gz"
  # version "VERSION_NUMBER"
  sha256 "SHA_256"

  def install
    bin.install "appsody"
    bin.install "appsody-controller"
    ohai "Checking prerequisites..."
    retval=check_prereqs
    if retval
      ohai "Done."
    else
      opoo "Docker not detected. Please ensure docker is installed and running before using appsody."
    end
  end

  def check_prereqs
    begin
      original_stderr = $stderr.clone
      original_stdout = $stdout.clone
      $stderr.reopen(File.new("/dev/null", "w"))
      $stdout.reopen(File.new("/dev/null", "w"))
      begin
        system("/usr/local/bin/docker", "ps")
        retval=true
      rescue
        retval=false
      end
    rescue => e
      $stdout.reopen(original_stdout)
      $stderr.reopen(original_stderr)
      raise e
    ensure
      $stdout.reopen(original_stdout)
      $stderr.reopen(original_stderr)
    end
    retval
  end

  test do
    raise "Test not implemented."
  end
end