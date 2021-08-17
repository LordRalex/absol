# Absol
Absol is a Discord bot created for many purposes, one of them being helping around in [Minecraft Community Support](https://discord.gg/58Sxm23).

Absol's main functions are: 
!?hjt (link) 
and
!?f or !?factoid (up to 5 factoids) (pings)

HJT or [HiJackThis](https://minecrafthopper.net/help/hjt/) is a tool used for obtaining diagnostic reports from your PC.

F or Factoid is used for sending large amounts of info in short snappy commands, you can find a list of factoids [here](https://cp.minecrafthopper.net/factoids).

## Building Absol
To build absol add 
```Dockerfile
ENV DISCORD_TOKEN="YOUR DISCORD BOT TOKEN"
ENV DATABASE=""
```
before
```Dockerfile
ENTRYPOINT ["/go/bin/absol"]
```
### Linux
Run `sudo docker build -t absol .` and if it dosent find any errors you should see `Successfully built`.
Then you can start the docker container by running `sudo docker run -it absol`

### Windows
Run `docker build -t absol .` and if it dosent find any errors you should see `Successfully built`.
Then you can start the docker container by running `docker run -it absol`

## Notices

* **Please note the Minecraft Community Support Server is NOT related to Mojang in any way!**

