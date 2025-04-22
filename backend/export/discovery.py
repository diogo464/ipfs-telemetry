from pydantic import BaseModel, Field, root_validator

DISCOVERY_SUBJECT = "discovery"


class DiscoveryNotification(BaseModel):
    id: str = Field(description="Peer ID of the discovered peer")
    addresses: list[str] = Field(
        description="List of multi addresses of the discovered peer"
    )

    # this allows piping `ipfs id` to curl, example:
    # ipfs id | curl -X POST -H 'Content-Type: application/json' --data @- http://localhost:8000/api/v1/discovery 
    # since the id command uses fields `ID` and `Addresses`
    @root_validator(pre=True)
    def normalize_case(cls, values):
        for key in ["ID", "Id"]:
            if key in values and "id" not in values:
                values["id"] = values[key]
        for key in ["Addresses"]:
            if key in values and "addresses" not in values:
                values["addresses"] = values[key]
        return values

