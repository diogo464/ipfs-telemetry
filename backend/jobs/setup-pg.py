import logging
import asyncio

import env
import db


async def main():
    conn_info = env.create_db_conn_info()
    engine = db.create_engine(conn_info)

    if db.requires_setup_database(engine):
        logging.info("Setting up database")
        db.setup_database(engine)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    asyncio.run(main())

