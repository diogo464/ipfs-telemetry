import logging

from fastapi import APIRouter
from rocketry import Grouper

logger = logging.getLogger(__name__)
router = APIRouter()
group = Grouper()
