require 'sinatra'
require 'logger'

configure do
  set :bind, '0.0.0.0'
  set :port, 8080
  disable :logging
  logger = Logger.new STDOUT
  logger.level = Logger::INFO
  logger.datetime_format = '%a %d-%m-%Y %H%M '
  set :logger, logger
end

get '/' do
  'Hello World, I\'m Ruby!'
end

after do
  # disable logging for kubernetes liveness/readiness probes
  if ENV.has_key? "KUBERNETES_SERVICE_HOST"
    if request.user_agent == 'Go-http-client/1.1'
      return
    end
  end
  settings.logger.info "#{response.status} #{response.body}"
end
