# Scorebot
Scorebot monitors the git repository for tags that look like `score(123)`.  
You can get the highest score by just saying the phrase "high score" on any channel that the @scorebot is invited to.
Each instance of the `score(x)` pattern is treated independently and will be shown as a separate entry on the scoreboard.

# Setup
There are two environment variables that need to be set whether you are running locally or in a container.
The `TARGET_REPO` is composed of the target repository and a working token to access that repo.
The `SLACKBOT_API_TOKEN` is gotten from the bot setup proces at: https://api.slack.com/bot-users

Locally:

    export TARGET_REPO=https://<GITHUB_TOKEN>@github.com/jemoster/icfp2017
    export SLACKBOT_API_TOKEN
    
In docker:

    --env TARGET_REPO=https://<MY_TOKEN>@github.com/jemoster/icfp2017 --env SLACKBOT_API_TOKEN
