from pydantic import BaseModel, Field

DISCOVERY_SUBJECT = "discovery"


class DiscoveryNotification(BaseModel):
    id: str = Field(description="Peer ID of the discovered peer")
    addresses: list[str] = Field(
        description="List of multi addresses of the discovered peer"
    )
