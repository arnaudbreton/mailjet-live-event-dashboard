# Mailjet Live Event API Dashboard

Demo application for [Mailjet](https://mailjet.com) [Event API](http://dev.mailjet.com/guides/event-api-guide/) built with [ReactJS](http://facebook.github.io/react/).
It's a live dashboard of all the event you receive via our API.
Based on offical [React tutorial](https://github.com/reactjs/react-tutorial)

## Demo

Go [here](http://37.187.61.17:3001/)

## Installation

Install [Golang](http://golang.org/) to run the API server.

Copy the `config.json.dist` file to `config.json` and fill it with your [API Keys](https://app.mailjet.com/account/api_keys) and [default sender](https://app.mailjet.com/account/sender).
Run the server: `go run server.go`. The server accepts an optional parameter to set the port.

Go to `localhost:port` (default port is 3000) and follow the instructions.

## Want to contribute? Need help?

Open a PR / issue here on Github.
Also, don't hesitate to email or [tweet](https://twitter.com/arnaud_breton) me!