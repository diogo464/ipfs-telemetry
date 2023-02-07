import datetime
import logging
import asyncio
import ipaddress

from typing import Optional
from geoip2.database import Reader
from nats.aio.client import Client as NatsClient
from nats.js.api import ConsumerConfig
from sqlalchemy.orm import Session
from sqlalchemy.exc import IntegrityError

import backend.monitor
import backend.db
import backend.env
import backend.discovery
import backend.util

logger = logging.getLogger(__name__)

CONSUMER_NAME = "exporter-pg"
QUEUE_NAME = CONSUMER_NAME


async def run():
    nc = await backend.env.create_nats_client()
    conn_info = backend.env.create_db_conn_info()
    engine = backend.db.create_engine(conn_info)
    done, _ = await asyncio.wait(
        [
            asyncio.create_task(
                _export_telemetry(nc, backend.db.create_session(engine))
            ),
            asyncio.create_task(
                _export_discovery(nc, backend.db.create_session(engine))
            ),
        ],
        return_when=asyncio.FIRST_COMPLETED,
    )
    for task in done:
        raise task.exception()


async def _export_telemetry(nc: NatsClient, session: Session):
    js = nc.jetstream()
    consumer = ConsumerConfig(
        durable_name=CONSUMER_NAME,
        description="Export telemetry to VictoriaMetrics",
    )
    sub = await js.subscribe(
        "telemetry.export", queue=QUEUE_NAME, config=consumer, stream="telemetry"
    )

    async for msg in sub.messages:
        try:
            export = backend.monitor.Export.parse_raw(msg.data)
            model_data = backend.db.convert_export(export)
            model_data.bulk_save(session)
            session.commit()
        except IntegrityError as e:
            session.rollback()
            logger.exception("Integrity error, skipping export", exc_info=True)
        await msg.ack()


async def _export_discovery(nc: NatsClient, session: Session):
    geodb = Reader("GeoLite2-City.mmdb")
    sub = await nc.subscribe("discovery", queue=QUEUE_NAME)
    async for msg in sub.messages:
        try:
            discovery = backend.discovery.DiscoveryNotification.parse_raw(msg.data)
            peer = discovery.id
            timestamp = datetime.datetime.utcnow()
            lat, lon, loc = _get_lat_lon_location_from_optional_ip(
                geodb, _get_first_public_ip_from_discovery(discovery)
            )
            model_data = backend.db.Discovery(
                peer=peer,
                timestamp=timestamp,
                latitude=lat,
                longitude=lon,
                location=loc,
            )
            with session.begin():
                session.add(model_data)
        except Exception as e:
            logger.exception(e)
            continue


def _get_lat_lon_location_from_optional_ip(
    geodb: Reader, ip: Optional[str]
) -> tuple[float, float, str]:
    try:
        response = geodb.city(ip or "")
        return (
            response.location.latitude or 0.0,
            response.location.longitude or 0.0,
            f"{response.country.name}/{response.city.name}",
        )
    except:
        return 0.0, 0.0, "unknown"


def _get_first_public_ip_from_discovery(
    discovery: backend.discovery.DiscoveryNotification,
) -> Optional[str]:
    for addr in discovery.addresses:
        if "circuit" in addr or "ip6" in addr:
            continue
        _, ip = backend.util.split_multiaddr_to_peer_and_ip(addr)
        ip_info = ipaddress.ip_address(ip)
        if ip_info.is_global:
            return ip
    return None
