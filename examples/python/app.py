import logging
import logging.config
import os
import sys
import time

from flask import Flask, request

RUNNING_IN_KUBERNETES = 'KUBERNETES_SERVICE_HOST' in os.environ

app = Flask(__name__)
handler = logging.StreamHandler(sys.stdout)
handler.setLevel(logging.INFO)
app.logger.addHandler(handler)

# disable flask's default request logger
werkzeug_logger = logging.getLogger('werkzeug')
werkzeug_logger.setLevel(logging.ERROR)

logging.getLogger().setLevel(logging.INFO)

@app.route('/', methods=['GET'])
def hello_world():
    return "Hello, World!\n"

@app.after_request
def log_after(response):
    # check if the request was made from a kubernetes liveness/readiness probe
    if RUNNING_IN_KUBERNETES:
        if request.headers.get('User-Agent') == 'Go-http-client/1.1':
            return response
    app.logger.info('%s %s %s %s %s %s',
        time.strftime('[%Y-%b-%d %H:%M]'),
        request.remote_addr,
        request.method,
        request.scheme,
        request.full_path,
        response.status)
    return response

if __name__ == '__main__':
    host = '0.0.0.0'
    port = int(os.environ.get('PORT', '8080'))
    app.logger.info('Serving on http://%s:%d', host, port)
    app.run(host=host, port=port)
