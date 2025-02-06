# twitch-ws
This is a simple application designed to relay twitch websocket information to all connected clients in a standard manner.

I wrote this because I've been utilizing the twitch websockets more in some personal projects, but having to port all the twitch connection logic into my different apps that are written in different languages is tedious. This allows me to have a single point of entry that has a standard I know about and can maintain if necessary.

>[!WARNING]
>This is not something that should be run with open access as it has no security features. This was written to run on my own internal network.

## How it works
To start, the server creates a connection to the Twitch EventSub websocket. It then subscribes to the events I care about.
A websocket server is then created that clients can connect to. The only requirement for this server is that a `ping` message is sent every 10 seconds and the clients should respond with a valid `pong` message within 10 seconds or they will be disconnected.

When a message is received from the Twitch EventSub, the data is massaged into the `models.Event` format, and then broadcast to all connected clients on the websocket server.

Messages from clients are dropped and not sent anywhere.
