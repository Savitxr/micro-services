FROM python:3.14.3-alpine@sha256:faee120f7885a06fcc9677922331391fa690d911c020abb9e8025ff3d908e510 AS base
FROM base AS builder
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
RUN apk update \
    && apk add --no-cache g++ linux-headers \
    && rm -rf /var/cache/apk/*
COPY requirements.txt .
RUN pip install -r requirements.txt
FROM base
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV ENABLE_PROFILER=1
RUN apk update \
    && apk add --no-cache libstdc++ \
    && rm -rf /var/cache/apk/*
WORKDIR /email_server
COPY --from=builder /usr/local/lib/python3.14/ /usr/local/lib/python3.14/
COPY . .
EXPOSE 8080
ENTRYPOINT [ "python", "email_server.py" ]
