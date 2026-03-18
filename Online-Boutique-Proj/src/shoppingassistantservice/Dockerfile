FROM python:3.14.3-slim@sha256:6a27522252aef8432841f224d9baaa6e9fce07b07584154fa0b9a96603af7456 AS base
FROM base AS builder
RUN apt-get -qq update \
    && apt-get install -y --no-install-recommends g++ \
    && rm -rf /var/lib/apt/lists/*
COPY requirements.txt .
RUN pip install -r requirements.txt
FROM base
ENV PYTHONUNBUFFERED=1
WORKDIR /shoppingassistantservice
COPY --from=builder /usr/local/lib/python3.14/ /usr/local/lib/python3.14/
COPY . .
ENV PORT "8080"
EXPOSE 8080
ENTRYPOINT ["python", "shoppingassistantservice.py"]
