## Subtitles

### Google Cloud credentials

1) Go to the Google Cloud Console (https://console.cloud.google.com/).
2) Create a new project or select an existing one.
3) Enable the necessary APIs:
In the left sidebar, click on "APIs & Services" > "Library".
Search for and enable the following APIs:
- Cloud Speech-to-Text API
- Cloud Translation API
4) Create a service account:
- In the left sidebar, click on "IAM & Admin" > "Service Accounts".
- Click "Create Service Account" at the top of the page.
- Enter a name for your service account and click "Create".
- For the role, select "Project" > "Owner" (or more restrictive roles if needed).
- Click "Continue" and then "Done".
5) Generate a key for the service account:
- In the service accounts list, find the account you just created.
- Click on the three dots in the "Actions" column and select "Manage keys".
- Click "Add Key" > "Create new key".
- Choose "JSON" as the key type and click "Create".
- The credentials file will be automatically downloaded to your computer.
- Rename the downloaded file to credentials.json and move it to the subtitles directory.