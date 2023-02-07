FROM docker.io/python:alpine3.17
LABEL org.opencontainers.image.source "https://github.com/diogo464/ipfs-telemetry"

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
COPY scripts/geolite_download.sh .
RUN sh geolite_download.sh
COPY poetry.lock pyproject.toml ./
RUN poetry install --only main
COPY . .

ENTRYPOINT [ "poetry", "run", "python", "-m", "backend", "--host", "0.0.0.0" ]