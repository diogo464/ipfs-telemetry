import logging

from fastapi import APIRouter
from rocketry import Grouper


import backend.env as env
import backend.discovery as discovery

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/discovery")
group = Grouper()


@router.post("")
async def discover(notif: discovery.DiscoveryNotification):
    logger.info(f"Received discovery notification: {notif}")
    nc = await env.create_nats_client()
    await nc.publish(discovery.DISCOVERY_SUBJECT, notif.json().encode("utf-8"))
