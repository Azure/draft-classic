/// <reference path="../node_modules/@types/node/index.d.ts" />

import * as http from 'http'

const port = 8080;

function requestHandler(request: http.IncomingMessage, response: http.ServerResponse): void {
    console.log(request.url);
    response.end("Hello World, I'm Node.js!");
}

const server = http.createServer(requestHandler);

server.listen(port, function (err: Error): void {
    if (err) {
        return console.log(err);
    }

    console.log(`server is listening on ${port}`);
})
