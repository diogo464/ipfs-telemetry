import re


def split_multiaddr_to_peer_and_ip(addr: str) -> tuple[str, str]:
    match = re.match(r"/ip4/(\d+\.\d+\.\d+\.\d+)/.*?/p2p/(\w+)", addr)
    if match is None:
        raise ValueError(f"Invalid multiaddr: {addr}")
    return match.group(2), match.group(1)
