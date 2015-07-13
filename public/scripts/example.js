/**
 * This file provided by Facebook is for non-commercial testing and evaluation purposes only.
 * Facebook reserves all rights not expressly granted.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * FACEBOOK BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
 * ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

var Event = React.createClass({
  render: function() {
    var rawMarkup = marked(this.props.children.toString(), {sanitize: true});
    return (
      <div className="Event">
        <h2 className="EventType">
          Event: {this.props.eventType}
        </h2>
        <span dangerouslySetInnerHTML={{__html: rawMarkup}} />
      </div>
    );
  }
});

var EventBox = React.createClass({
  loadEventsFromServer: function() {
    $.ajax({
      url: this.props.url,
      dataType: 'json',
      cache: false,
      success: function(data) {
        this.setState({data: data});
      }.bind(this),
      error: function(xhr, status, err) {
        console.error(this.props.url, status, err.toString());
      }.bind(this)
    });
  },
  handleEventSubmit: function(Event) {
    $.ajax({
        url: this.props.submitUrl,
        dataType: 'json',
        type: 'POST',
        data: JSON.stringify(Event),
        success: function(data) { // callback method for further manipulations             
          console.log(data);
        },
        error: function(xhr, status, err) {
          console.error(this.props.submitUrl, status, err.toString());
        }.bind(this)
      });
  },
  getInitialState: function() {
    return {data: []};
  },
  componentDidMount: function() {
    this.loadEventsFromServer();
    setInterval(this.loadEventsFromServer, this.props.pollInterval);
  },
  render: function() {
    return (
      <div className="EventBox">
        <h1>Mailjet Live Event API Dashboard</h1>
        <EventList data={this.state.data} />
        <SendForm onEventSubmit={this.handleEventSubmit} />
      </div>
    );
  }
});

var EventList = React.createClass({
  render: function() {
    var EventNodes = this.props.data.map(function(event, index) {
      console.log(event)
      return (
        // `key` is a React-specific concept and is not mandatory for the
        // purpose of this tutorial. if you're curious, see more here:
        // http://facebook.github.io/react/docs/multiple-components.html#dynamic-children
        <Event eventType={event.EventType} key={index}>
          {event.MessageID}
        </Event>
      );
    });
    return (
      <div className="EventList">
        {EventNodes}
      </div>
    );
  }
});

var SendForm = React.createClass({
  handleSubmit: function(e) {
    e.preventDefault();
    var apiKey = React.findDOMNode(this.refs.apiKey).value.trim();
    var apiSecret = React.findDOMNode(this.refs.apiSecret).value.trim();
    var fromEmail = React.findDOMNode(this.refs.fromEmail).value.trim();
    var recipient = React.findDOMNode(this.refs.recipient).value.trim();
    var subject = React.findDOMNode(this.refs.subject).value.trim();
    var body = React.findDOMNode(this.refs.body).value.trim();
    if (!recipient || !body) {
      return;
    }
    this.props.onEventSubmit({
      ApiKey: apiKey,
      ApiSecret: apiSecret,
      FromEmail: fromEmail,
      Recipient: recipient, 
      Subject: subject, 
      Body: body
    });
    React.findDOMNode(this.refs.recipient).value = '';
    React.findDOMNode(this.refs.subject).value = '';
    React.findDOMNode(this.refs.body).value = '';
  },
  render: function() {
    return (
      <form className="SendForm" onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label htmlFor="apiKey">Mailjet API Key:</label>
          <input type="text" id="apiKey" ref="apiKey" className="form-control"/>
          <label htmlFor="apiSecret">Mailjet API Secret:</label>
          <input type="text" id="apiSecret" ref="apiSecret" className="form-control"/>
        </div>
        <div className="form-group">
          <label htmlFor="fromEmail">From email (valid Mailjet sender)</label>
          <input type="text" placeholder="api@mailjet.com" id="fromEmail" ref="fromEmail" className="form-control"/>
          <label htmlFor="recipient">To</label>
          <input type="text" placeholder="api@mailjet.com" id="recipient" ref="recipient" className="form-control"/>
          <label htmlFor="subject">Subject</label>
          <input type="text" placeholder="Hello World!" id="subject" ref="subject" className="form-control"/>
          <label htmlFor="body">Subject</label>
          <textarea placeholder="Say something..." id="body" ref="body" className="form-control"></textarea>
          <input type="submit" value="Send!" className="btn btn-default"/>
        </div>
      </form>
    );
  }
});

React.render(
  <EventBox url="events.json" submitUrl="messages" pollInterval={2000} />,
  document.getElementById('content')
);
