# Go Setup Instructions

## Install Go

## Download the Go Google Sheets API packages.
go get -u google.golang.org/api/sheets/v4
go get -u golang.org/x/oauth2/google


### Setup Google Cloud Project (via IAM & Admin)

### Setup credentials and OAuth Login Screen

### Download JSON credentials file to target machine


# Running the Application

When running for the first time (or the first time in a while), app will present a URL to 
paste into the browser to permit the app access to the spreadsheets. After consenting access,
the webpage will redirect to localhost, with an auth code in the URL:

      http://localhost/?state=state-token&code=<AUTHCODE>&scope=https://www.googleapis.com/auth/spreadsheets.readonly

Paste this auth code into the console where the app is running to authorize the application and continue.

