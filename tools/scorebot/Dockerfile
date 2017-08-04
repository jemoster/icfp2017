FROM python:3

MAINTAINER jemoster@gmail.com

RUN apt-get install -y git

ADD requirements.txt /
RUN pip install -r requirements.txt
ADD *.py /

CMD [ "python", "./main.py" ]
