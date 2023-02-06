from __future__ import annotations

import datetime
import time

from pydantic import BaseModel
from typing import Callable, Generator, Union
from minio import Minio

import lib.env

TELEMETRY_CONTENT_TYPE = "application/json"


def _datetime_to_filename(dt: datetime.datetime) -> str:
    """
    Convert a datetime object to a filename.
    """
    return f"{dt.year}/{dt.month}/{dt.day}/{dt.hour}/{dt.minute}/{dt.second}/{int(dt.timestamp() * 1000 // 1)}"


def _datetime_to_millis(dt: datetime.datetime) -> int:
    """
    Convert a datetime object to a millisecond timestamp.
    """
    return int(dt.timestamp() * 1000 // 1)


def _filename_to_milliseconds(filename: str) -> int:
    """
    Convert a filename to a millisecond timestamp.
    """
    return int(filename.split("/")[-1].split(".")[0])


def _filename_to_datetime(filename: str) -> datetime.datetime:
    """
    Convert a filename to a datetime object.
    """
    return datetime.datetime.utcfromtimestamp(
        _filename_to_milliseconds(filename) / 1000
    )


def _object_id_to_datetime(object_id: str) -> datetime.datetime:
    """
    Convert a object_id to a datetime object.
    """
    return datetime.datetime.utcfromtimestamp(int(object_id) / 1000)


class TelemetryObjectId(str):
    @classmethod
    def __get_validators__(cls) -> Generator[Callable, None, None]:
        yield cls.validate

    @classmethod
    def validate(cls, v: Union[str, int, datetime.datetime]) -> TelemetryObjectId:
        if isinstance(v, str):
            last = v.split("/")[-1]
            if last.isdigit():
                return TelemetryObjectId(last)
            raise ValueError(f"Invalid TelemetryObjectId: {v}")
        if isinstance(v, int):
            return TelemetryObjectId(str(v))
        if isinstance(v, datetime.datetime):
            return TelemetryObjectId(str(_datetime_to_millis(v)))
        raise ValueError(f"Invalid TelemetryObjectId: {v}")

    @property
    def filename(self) -> str:
        return _datetime_to_filename(_object_id_to_datetime(self))

    @staticmethod
    def from_datetime(dt: datetime.datetime) -> TelemetryObjectId:
        return TelemetryObjectId(dt)

    @staticmethod
    def from_filename(filename: str) -> TelemetryObjectId:
        return TelemetryObjectId(filename.split("/")[-1])

    @staticmethod
    def from_object_id(object_id: str) -> TelemetryObjectId:
        return TelemetryObjectId(object_id)


class TelemetryObject(BaseModel):
    object_id: TelemetryObjectId
    data: bytes


def optional_dt_to_toid(
    dt: Union[datetime.datetime, TelemetryObjectId, None]
) -> Union[TelemetryObjectId, None]:
    """
    Convert a datetime object or TelemetryObjectId to a TelemetryObjectId.
    """
    if dt is None:
        return None
    if isinstance(dt, TelemetryObjectId):
        return dt
    return TelemetryObjectId.from_datetime(dt)


def stream_telemetry_object_ids(
    minio: Minio,
    start: Union[datetime.datetime, TelemetryObjectId, None] = None,
    end: Union[datetime.datetime, TelemetryObjectId, None] = None,
    bucket: str = lib.env.S3_BUCKET_TELEMETRY,
    interval: datetime.timedelta = datetime.timedelta(minutes=1),
    stream: bool = True,
) -> Generator[TelemetryObjectId, None, None]:
    """
    Yield a list of filenames for telemetry objects in the given range.
    """
    toid_start = optional_dt_to_toid(start)
    toid_end = optional_dt_to_toid(end)

    if toid_start is not None and toid_end is not None and toid_start > toid_end:
        raise ValueError("start must be before end")

    start_after = None if toid_start is None else toid_start.filename
    while True:
        objects = minio.list_objects(bucket, recursive=True, start_after=start_after)
        for obj in objects:
            toid = TelemetryObjectId.from_filename(obj.object_name)
            if toid_end is not None and toid > toid_end:
                break
            if toid_start is None or toid >= toid_start:
                toid_start = toid
            yield toid
        if not stream:
            break
        time.sleep(interval.total_seconds())


def download_telemetry_object(
    minio: Minio,
    toid: TelemetryObjectId,
    bucket: str = lib.env.S3_BUCKET_TELEMETRY,
) -> TelemetryObject:
    """
    Download a telemetry object from S3.
    """
    data = minio.get_object(bucket, toid.filename).data
    return TelemetryObject(object_id=toid, data=data)
