import json

def move():
    return json.dumps({
    "eth0": {
        "ip": "11.22.194.72",
        "mac": "00:ab:f7:e4:74:c2",
        "meta": {
            "type": "ethernet-interface"
        },
        "name": "eth0",
        "netmask": "255.255.248.0",
        "network": "11.22.192.0",
        "status": "up"
    }
})

