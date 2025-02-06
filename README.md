# twitch-ws
This is a simple application designed to relay twitch websocket information to all connected clients in a standard manner.

I wrote this because I've been utilizing the twitch websockets more in some personal projects, but having to port all the twitch connection logic into my different apps that are written in different languages is tedious. This allows me to have a single point of entry that has a standard I know about and can maintain if necessary.

>[!WARNING]
>This is not something that should be run with open internet access as it has no security features. This was written to run on my own internal network.

## Running
1. Before you start, you will need to go to the [Twitch Developer Console](https://dev.twitch.tv/console) and register a new application. The oauth redirect URL should be `http://localhost:7000` and make sure the client type is *Confidential*
1. After creating your application, copy your Client ID as we will use it later
1. Get your Twitch User ID. If you don't know it, you can use this website: https://www.streamweasels.com/tools/convert-twitch-username-%20to-user-id/
1. Create a new file named `.env` and paste the following contents, where `<CLIENT_ID>` is the value copied from step 2 and `<USER_ID>` is the value copied from step 3
```
TWITCH_CLIENT_ID=<CLIENT_ID>
TWITCH_USER_ID=<TWITCH_USER_ID>
```
5. Run the server using `go run cmd/server.go`. It should fail, but should also provide a link to open.
6. Open the link, and copy the accessToken parameter out of the URL. Add this to your `.env` file in the following format:
```
TWITCH_ACCESS_TOKEN=<ACCESS_TOKEN>
```
7. Run the app again, and it should successfully register for events.
8. Any clients you want to get the receive message should connect to `ws://localhost:8080/ws`.
9. Once connected, anytime your client receives a `PING` message, they should respond with a valid `PONG` message or they will be disconnected. 

## How it works
To start, the server creates a connection to the Twitch EventSub websocket. It then subscribes to the events I care about.
A websocket server is then created that clients can connect to. The only requirement for this server is that a `ping` message is sent every 10 seconds and the clients should respond with a valid `pong` message within 10 seconds or they will be disconnected.

When a message is received from the Twitch EventSub, the data is massaged into the `models.Event` format, and then broadcast to all connected clients on the websocket server.

Messages from clients are dropped and not sent anywhere.
