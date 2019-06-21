class Appsody < Formula
    desc "Appsody command-line interface"
    homepage "https://github.com/appsody/appsody#readme"
    version "VERSION_NUMBER"
    url "https://github.com/REPO_NAME/releases/download/RELEASE_TAG/FILE_PREFIX-VERSION_NUMBER.tar.gz" 
    sha256 "SHA_256"
  #  url "https://github.com/appsody/appsody/archive/master.zip",
  #      :tag      => "master"
  #      :revision => "b5d4f876f4bfe294b70c092210613e30eceb0dc4"
  #  head "git@github.com:appsody/appsody.git"
  
  #  bottle do
  def install
    bin.install "appsody"
    bin.install "appsody-controller" 
    ohai "Checking prerequisites..."
    #system "docker"
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
      $stderr.reopen(File.new('/dev/null', 'w'))
      $stdout.reopen(File.new('/dev/null', 'w'))
      begin
        system('/usr/local/bin/docker', 'ps')
        retval=true
      rescue
        retval=false
      end
    rescue Exception => e
      $stdout.reopen(original_stdout)
      $stderr.reopen(original_stderr)
      raise e
    ensure
      $stdout.reopen(original_stdout)
      $stderr.reopen(original_stderr)
    end
    return retval
  end  

  
    test do
      raise "Test not implemented."
    end
  end
  
