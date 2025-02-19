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

## Testing
It is advised to do your testing against a demo server. This can be done using the [twitch-cli](https://dev.twitch.tv/docs/cli/) and any other sample websocket app (like https://piehost.com/websocket-tester)
1. To start, enable the Twitch test server by running `twitch-cli event websocket start-server -S`
2. In your `.env` file, set `TWITCH_WEBSOCKET_URL=ws://127.0.0.1:8080/ws` and set `TWITCH_SUBSCRIPTION_URL=http://127.0.0.1:8080/eventsub/subscriptions`
3. Make sure in your `.env` file that if you have `WS_PORT` set, it is _not_ set to `8080` or it will conflict with the Twitch websocket server
4. Run the app, using the compiled binary or by running `go run cmd/server.go`
5. Once the connection is validated, use your preferred websocket client to connect to `ws://localhost:8000/ws`
6. In a separate window, you can now trigger `twitch-cli` commands to see how they look on the receiving client.
    - Make sure to use `--transport=websocket` option
    - IE: `twitch-cli event trigger channel.cheer --transport=websocket`

## How it works
To start, the server creates a connection to the Twitch EventSub websocket. It then subscribes to the events I care about.
A websocket server is then created that clients can connect to. The only requirement for this server is that a `ping` message is sent every 10 seconds and the clients should respond with a valid `pong` message within 10 seconds or they will be disconnected.

When a message is received from the Twitch EventSub, the data is massaged into the `models.Event` format, and then broadcast to all connected clients on the websocket server.

Messages from clients are dropped and not sent anywhere.

### Message format
All messages will be formatted into the following, with appropriate values based on the event type (IE: for a `subscription` event, only the `subscription` subsection will be available)

If an action includes a message (such as a reward redemption, subscription, bits, etc) then that message will always be available in the root `message` value.
```jsonc
{
    "type": "", // See the list of types outlined below
    "display_name": "", // The Twitch display name of the user who performed the action
    "user_id": "", // The Twitch user ID of the user who performed the action
    "message": "", // If a message is included from the user
    "is_anonymous": "", // If the requested this action anonymously
    "subscription": {
        "from_user_id": "", // If gifted, the user ID of that person
        "from_user_display_name": "", // If gifted, the display name of that person
        "is_gift": false, // Is this a gifted subscription
        "tier": "", // The tier level of the subscription. Note that this will be 1000, 2000, or 3,000
        "total": 0, // How many months the user has subbed for in total
    },
    "channel_point_redemption": {
        "reward_id": "", // The ID of the reward that was redeemed
        "title": "", // The title of the reward that was redeemed
        "cost": 0, // The cost of the reward that was redeemed
        "prompt": "" // If there was a prompt from the reward (note that the user message will be in the root `message` field
    },
    "bits": {
        "amount": 0 // The amount of bits spent
    },
    "raid": {
        "viewers": 0 // The number of viewers brought along with the raid
    }
}
```

### Supported types

| EventSub Type | Local Type | Description | Notes |
| --- | --- | --- | --- |
| `channel.follow` | `follow` | A user follows the channel | | 
| `channel.subscription.gift` | `gift_sub` | A user is gifted a sub | `is_gift` will be `true` | 
| `channel.subscribe` | `subscribe` | A user subscribes | `is_gift` will be `false` | 
| `channel.subscription.message` | `resubscribe` | When a user resubscribes and includes a message, this is the type | | 
| `channel.chat.message` | `chat` | A chat message is sent | | 
| `channel.channel_points_custom_reward_redemption.add` | `channel_points` | When a user redeems a custom channel point | | 
| `channel.cheer` | `bits` | When a user spends bits | | 
| `channel.raid` | `raid` | When a user raids the watched channel | | 
