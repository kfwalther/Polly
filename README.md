# Polly Portfolio Tracker

This portfolio tracking application was implemented using [**Go**](https://go.dev/) and [**React**](https://reactjs.org/). Please use the instructions below to setup these tools for development.

## Setup Go

Install Go by downloading the installer from [**here**](https://go.dev/dl/). Once installed, build the backend Go server for the application:

    cd backend
    go build -v -o go-server.exe

Download the JSON credentials file to target machine (if credentials.json is not already present). You can download the JSON file from the Google Cloud API [**Credentials**](https://console.cloud.google.com/apis/credentials) page, under the **OAuth 2.0 Client IDs** section.

### Run the backend server

Run the `go-server.exe` executable from the command line. When running for the first time (or the first time in a while), app will present a URL to
paste into the browser to permit the app access to the Google Sheets. After consenting access, the webpage will redirect to localhost, with an auth code in the URL:

    http://localhost/?state=state-token&code=<AUTHCODE>&scope=https://www.googleapis.com/auth/spreadsheets.readonly

Paste this AUTHCODE into the console where the app is running to authorize the application and continue.

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


