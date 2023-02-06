import uvicorn
import asyncio
import logging
import argparse

from fastapi import FastAPI
from rocketry import Rocketry
from signal import SIGINT, SIGTERM

from .routers import raw, discovery
from .tasks import exporter_vm, exporter_pg


logger = logging.getLogger(__name__)


class Service(uvicorn.Server):
    def __init__(self, config: uvicorn.Config, rocket: Rocketry):
        super().__init__(config)
        self.rocket = rocket

    def handle_exit(self, sig: int, frame) -> None:
        # https://rocketry.readthedocs.io/en/stable/cookbook/fastapi.html
        self.rocket.session.shut_down()
        return super().handle_exit(sig, frame)


async def exception_wrapper(coroutine):
    try:
        return await coroutine
    except Exception as e:
        logger.fatal("Exception in task", exc_info=True)
        raise e


async def main(host: str, port: int, root_path: str):
    app = FastAPI(root_path=root_path)
    app.include_router(raw.router)
    app.include_router(discovery.router)

    rocket = Rocketry(execution="async")
    rocket.include_grouper(raw.group)
    rocket.include_grouper(discovery.group)

    server = Service(
        config=uvicorn.Config(
            app,
            loop="asyncio",
            log_level="info",
            host=host,
            port=port,
            root_path=root_path,
        ),
        rocket=rocket,
    )
    couroutines = [
        server.serve(),  # web server
        rocket.serve(),  # rocketry
        # tasks
        exporter_vm.run(),
        exporter_pg.run(),
    ]

    async_tasks = [asyncio.create_task(exception_wrapper(c)) for c in couroutines]
    _, pending = await asyncio.wait(async_tasks, return_when=asyncio.FIRST_COMPLETED)
    for task in pending:
        task.cancel()


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)

    parser = argparse.ArgumentParser()
    parser.add_argument("--host", default="127.0.0.1", help="Address to bind to")
    parser.add_argument("--port", default=8000, type=int, help="Port to bind to")
    parser.add_argument("--root-path", default="/", help="Root path to serve on")
    args = parser.parse_args()

    # https://stackoverflow.com/questions/48562893/how-to-gracefully-terminate-an-asyncio-script-with-ctrl-c
    loop = asyncio.get_event_loop()
    main_task = asyncio.ensure_future(main(args.host, args.port, args.root_path))
    for signal in [SIGINT, SIGTERM]:
        loop.add_signal_handler(signal, main_task.cancel)
    try:
        loop.run_until_complete(main_task)
    finally:
        loop.close()
