FROM golang:1.18.0-alpine3.15

# Install python/pip
ENV PYTHONUNBUFFERED=1
RUN apk add --update --no-cache python3 && ln -sf python3 /usr/bin/python
RUN python3 -m ensurepip
RUN pip3 install --no-cache --upgrade pip setuptools

# Install required packages
WORKDIR /tmp
COPY tests/requirements.txt requirements.txt
RUN pip3 install -r requirements.txt

COPY --from=chat-client:latest /client /client

WORKDIR /tests
COPY tests/* ./

CMD [ "pytest", "-rP" ] 