# Mailjet Live Event API Dashboard

Demo application for [Mailjet](https://mailjet.com) [Event API](http://dev.mailjet.com/guides/event-api-guide/) built with [ReactJS](http://facebook.github.io/react/).
It's a live dashboard of all the event you receive via our API.
Based on the offical [React tutorial](https://github.com/reactjs/react-tutorial)

## Demo

Go [here](http://37.187.61.17:3001/)

## Installation

Install [Golang](http://golang.org/) to run the API server.
Install [Bower](http://bower.io/), the package manager for the web.

Install server dependencies: 

Copy the `config.json.dist` file to `config.json` and fill it with the following information (all optional):
*`base_url`: the Mailjet base URL. Default to our production environment
*`max_events_count': the maximum number of events to display, default to 10. 0 for unlimited
*`api_key`: the default Mailjet API key to use
*`api_secret`: the default Mailjet API secret to use
*`recipient`: the default email address to send the sample email to
*`subject`: the default subject of the sample email
*`body`: the default body of the sample email

Run the bower install command to fetch front-end dependencies: `bower install`
Run the server: `go run server.go`. The server accepts an optional parameter to set the port.
Go to `localhost:port` (default port is 3000) and follow the instructions.

## Want to contribute? Need help?

Open a PR / issue here on Github.
Also, don't hesitate to email or [tweet](https://twitter.com/arnaud_breton) me!

## (Low) hanging fruits
*Rewrite the API in React, to make it isomorphic
*Improve the design