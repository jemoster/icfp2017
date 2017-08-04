import re
import os

import git
from git import Repo
from slackbot.bot import Bot
from slackbot.bot import listen_to


score_regex = re.compile(r"score\([eE\d.-]+\)")

def get_score(tag):
    try:
        score = score_regex.match(str(tag)).group(0)
        if not score:
            return
        return float(score[6:-1])
    except:
        pass


def get_scoreboard():
    repo = Repo("icfp2017")
    assert not repo.bare

    scoreboard = []

    for tag in repo.tags:
        score = get_score(tag)
        if not score:
            continue
        scoreboard.append(
            (score, str(tag.commit))
        )

    return sorted(scoreboard, key=lambda x: x[0], reverse=True)

def update_tags():
    repo = Repo("icfp2017")

    #clear old tags
    for tag_reg in repo.tags:
        if score_regex.search(str(tag_reg)):
            repo.delete_tag(tag_reg)

    for remote in repo.remotes:
        remote.fetch()

def setup_repo():
    try:
        repo = Repo("icfp2017")
    except git.exc.NoSuchPathError:
        git.Git().clone(os.environ['TARGET_REPO'], 'icfp2017')


@listen_to('high score', re.IGNORECASE)
def hi(message):
    update_tags()
    scoreboard = get_scoreboard()
    if not scoreboard:
        message.reply("I don't have any scores to report yet")
    else:
        message.reply("High score: {} at https://github.com/jemoster/icfp2017/commit/{}".format(*scoreboard[0]))


def main():
    setup_repo()
    bot = Bot()
    bot.run()

if __name__ == "__main__":
    main()
