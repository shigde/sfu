# Shig WebClient

The Shig Web Client is a standalone version for a Shig instance. Shig is intended to act as an intermediary between Fediverse instances. The client provides the possibility to directly enter a Shig Lobby. Since Shig itself is not a streaming service, unfortunately, it is not yet possible to create a new live stream via the Shig Client.
Nevertheless, it is possible to start an existing stream live.
The stream must have been created beforehand as an interactive live stream through PeerTube or Owncats.

## Development

```shell
npm install
npm start
```

Navigate to `http://localhost:4200/`.


## Shig Instance as a Dependency

The web client essentially serves as a frontend for a Shig instance; therefore, it requires one. 

Please clone the sfu project. https://github.com/shigde/sfu

Normally, the SFU (Shig instance) listens for data from other Fediverse instances. 
However, to facilitate development, please start the SFU with an included SQLite database. 
Run inside the SFU project:

```shell
mv shig-dev.db shig.db
```
This will place a database with date in the root directory. 
After this launch the SFU (the Shig instance) using 

```shell
make run
```

Now you can lunch the Shig Web Client


## Develop with shig js sdk

### Setup, Build and Link the shig-js-sdk

```shell
git clone https://github.com/shigde/shig-js-sdk.git
cd shig-js-sdk
npm run watch
cd dist/core
npm link
```

### Setup project with linked shig-js-sdk

The Shig WebClient is a good place to develop the SDK as well.
To set up a develop environment for the SDK fallow the next steps.

```shell
cd web-client
npm i @shigde/core
npm start
```

Have Fun!



