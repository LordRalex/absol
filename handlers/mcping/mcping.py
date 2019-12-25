#!/usr/bin/env python
from mcstatus import MinecraftServer
import sys

try:
    server = MinecraftServer.lookup(sys.argv[1])
    status = server.status()
    print("The server is running {2}, has {0} players, replied in {1} ms".format(status.players.online, status.latency, status.version.name))
except Exception as e:
    print("Error: {0}".format(str(e)))