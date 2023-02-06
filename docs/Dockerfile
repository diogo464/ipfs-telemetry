FROM docker.io/python:alpine3.17

# Configure Poetry
ENV POETRY_VERSION=1.3.2
ENV POETRY_HOME=/opt/poetry
ENV POETRY_VENV=/opt/poetry-venv
ENV POETRY_CACHE_DIR=/opt/.cache

# Install poetry separated from system interpreter
RUN python3 -m venv $POETRY_VENV \
    && $POETRY_VENV/bin/pip install -U pip setuptools \
    && $POETRY_VENV/bin/pip install poetry==${POETRY_VERSION}

# Add `poetry` to PATH
ENV PATH="${PATH}:${POETRY_VENV}/bin"

WORKDIR /usr/src/app
COPY poetry.lock pyproject.toml ./
RUN poetry install --only main
COPY . .
RUN poetry run mkdocs build

FROM docker.io/nginx:1.21.4-alpine

LABEL org.opencontainers.image.source "https://github.com/diogo464/ipfs_telemetry"

COPY --from=0 /usr/src/app/site /usr/share/nginx/html