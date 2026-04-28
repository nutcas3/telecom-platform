Gem::Specification.new do |spec|
  spec.name          = "telecom-sdk"
  spec.version       = "1.0.0"
  spec.authors       = ["Telecom Platform Team"]
  spec.email         = ["team@telecom.example.com"]
  spec.summary       = "Ruby SDK for Telecom Platform"
  spec.description   = "Ruby SDK for interacting with the Telecom Platform API"
  spec.homepage      = "https://github.com/nutcas3/telecom-platform"
  spec.license       = "MIT"
  spec.required_ruby_version = ">= 3.0.0"

  spec.files = Dir.chdir(File.expand_path(__dir__)) do
    `git ls-files -z`.split("\x0").reject { |f| f.match(%r{\A(?:test|spec|features)/}) }
  end
  
  spec.bindir        = "exe"
  spec.executables   = spec.name.gsub(/^.*:/, "")
  spec.require_paths = ["lib"]

  spec.add_dependency "httparty", "~> 0.23"
  spec.add_dependency "websocket-client-simple", "~> 0.4"
  spec.add_dependency "json", "~> 2.7"
  
  spec.add_development_dependency "minitest", "~> 5.25"
  spec.add_development_dependency "rubocop", "~> 1.72"
end
