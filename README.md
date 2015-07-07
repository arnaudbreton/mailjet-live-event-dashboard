# Mailjet Live Event API Dashboard

A live dashboard built in [ReactJS](http://facebook.github.io/react/), for [Mailjet](https://mailjet.com) [Event API](http://dev.mailjet.com/guides/event-api-guide/).
Based on offical [React tutorial](https://github.com/reactjs/react-tutorial)

## Demo

Go [here]()

## Installation

Install [Golang](http://golang.org/) to run the server.

Copy the `config.json.dist` file to `config.json` and fill it with your [API Keys](https://app.mailjet.com/account/api_keys) and [default sender](https://app.mailjet.com/account/sender).
Run the server: `go run server.go`. The server accepts an optional parameter to set the port.

Setup a callback URL for [Mailjet Event](https://app.mailjet.com/account/triggers) or via our [Event API](http://dev.mailjet.com/guides/event-api-guide/).
Go to `localhost:port` and send an email via the form.

Subscribed event should appear for the message sent.