# Research

## Legacy Server List Ping

I've determined that the pre-netty `0xFE 0x01` server list ping packet (for Minecraft 1.4) first appeared in snapshot `12w42b`. I'm very confident that this is the case, but there is no guarantee. Here is my research:

 - [This snapshot](https://minecraft.gamepedia.com/Java_Edition_12w42b) was released on 18/10/2012. The wiki.vg pre-release protocol specification [14/10/2020 revision](https://wiki.vg/index.php?title=Pre-release_protocol&oldid=2908) did not contain the `0xFE 0x01` packet, whereas the [18/10/2020 revision](https://wiki.vg/index.php?title=Pre-release_protocol&oldid=2909) did.
 - The [Mojang article](https://web.archive.org/web/20161220125143/http://www.mojang.com/2012/10/minecraft-snapshot-12w42b-preprerelease/) for this snapshot mentioned that "server ping will now tell if the server and client have mismatching protocol versions". This was not mentioned in their [article for the previous snapshot](https://web.archive.org/web/20181203152033/https://www.mojang.com/2012/10/minecraft-snapshot-12w42a/).

The wiki.vg [protocol version number reference](https://wiki.vg/Protocol_version_numbers#Versions_after_the_Netty_rewrite) mentions the protocol version number for snapshot `12w42b` is 46 (along with `12w42a`, `12w41b` and `12w41a`). It might be more accurate to consider version 47 to be the protocol for this change, but nonetheless I will be chosing 46 (please open an issue if I am incorrect).

This maps to `002E` in hex, which is the number I will use in this package.