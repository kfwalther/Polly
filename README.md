![go](https://img.shields.io/badge/Golang-green?style=flat&logo=Go&label=Web%20Server&link=https%3A%2F%2Fgo.dev%2F)
![python](https://img.shields.io/badge/Python3-yellow?style=flat&label=Scripts&link=https%3A%2F%2Fwww.python.org%2Fdownloads%2F)
![react](https://img.shields.io/badge/React-lightblue?style=flat&label=Frontend&link=https%3A%2F%2Freactjs.org%2F)

# Polly Portfolio Tracker

This portfolio tracking application was implemented using [**Go**](https://go.dev/) and [**React**](https://reactjs.org/). Please use the instructions below to setup these tools for development.

## Setup Go

Install Go by downloading the installer from [**here**](https://go.dev/dl/). Once installed, build the backend Go server for the application:

    cd backend
    go build -v -o go-server.exe

Download the JSON credentials file to target machine (if credentials.json is not already present). You can download the JSON file from the Google Cloud API [**Credentials**](https://console.cloud.google.com/apis/credentials) page, under the **OAuth 2.0 Client IDs** section.

### Install and setup MongoDB

The web backend depends on a MongoDB database to store the wealth of information pulled from Yahoo Finance. Download and install MongoDB (Community Edition) for Windows [**here**](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-on-windows/). Once installed, start MongoDBCompass, connect to the MongoDB server, and create a new time-series database named `polly-data-prod`.

### Install Python 3

Python 3 is used as a helper script for querying stock data from Yahoo finance. Install Python from [**here**](https://www.python.org/downloads/).

### Run the backend server

Run the `go-server.exe` executable from the command line. When the server is finished querying data from Yahoo finance and reading Google sheets, it will begin listening for incoming requests from the frontend.

## Setup NPM & React

1. Download and install Node and NPM [**here**](https://docs.npmjs.com/cli/v7/configuring-npm/install).
2. Ensure Python 3 is installed and available in the `PATH`.
3. Ensure Visual Studio is installed with the following workloads:
    * Desktop Development with C++
    * Windows 10 SDK
4. Navigate to the `ui` directory and install the npm packages for this application:

       cd ui
       npm install

### Start the frontend

To start the React web server, run the following from the `ui` directory:

    npm start

Once started, you should be able to access the app in the browser here:

    http://localhost:3000
