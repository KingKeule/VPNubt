# VPNubt (VPN-udp-broadcast-tunneler)
Our tool "copies" udp broadcasts on the selected port to udp unicasts which are sent to the specific IP adress to bypass the VPN router barrier.

## Background
We love to play old school games like Warcraft 3 with friends. 
Since we can't do a LAN session like in our youth, we play over the internet via VPN without using Battle.Net.
The problem with e.g. Warctaft 3 is that the server could not be found even if we are connected via VPN.
> (VPN means here classic OSI layer 3 VPNs and not a OSI layer 2 bridge VPN.) 

## What is the reason for that?
The game server sends an udp broadcast to notify all player in the LAN. When you play over internet via VPN there is normaly a consumer router which do not relay this braodcast otherwise the network/internet would be flooded.
> (Only professional routers could do this with a directed broadcast)

## How we solved the problem
We have programmed a tool that listen on the selected network interface for udp broadcasts. If an udp broadcast is detected, its payload is copied into an udp unicast packet and then sent to the VPN receiver, because a unicast is not filtered by the router.

## Are there other solutions for this problem? 
All of the following tools solve the problem, but in a different way. They do not "convert" the broadcast and instead send an fixed predefined communication specifically for Warcraft 3.
* [LanCraft](https://gaming-tools.com/warcraft-3/lancraft/) (updated in 2008)
* [WC3Proxy](http://lancraft.blogspot.com/p/wc3proxy.html) ([Sourcecode @ GitHub](https://github.com/evshiron/wc3proxy), updated in 2015)
* [YAWLE](http://lancraft.blogspot.com/2008/08/yawle-yet-another-warcarft-lan-emulator.html) (updated in 2008)

## Why a new tool? 
Some of the programs mentioned above only work specifically for one game. 
Our tool on the other hand can be used universally and is not limited to Warcraft 3, for example.
In addition, we wanted to realize the implementation in a current programming language (GO).

- - - -

# Reverse enigneering (of Warcraft 3)
If you want to know how we reengineered it, read on here.

1. Identify the communication port of the game (on Windows 10)
   * Start the game (Warcraft 3) and entert he multiplayer lobby
   * Switch to windows and open the command line and type: ***tasklist | findstr war3.exe***
   * Note the displayed process id of Warcraft 3
   * Type in command line: ***netstat -ano | findstr <Warcraft 3 process id>***
   * So finally we find out that Warcraft is listen only for ***UDP communication on port 6112***
  
  
2. Understand the Warcraft 3 communication on UDP port 6112
   * Install and start [Wireshark](https://www.wireshark.org/)
   * Set the Wireshark displayfilter to: ***udp.port == 6112***
   * You can divide it in 3 Parts:

     1. **"Hello" information"**  
     When you enter the network lobby, Warcraft will only send a notifcation boradcast ***once***:
        * Source: _local IP of client_
        * Destination: _255.255.255.255_
        * Port: _UDP 6112_
        * Data: _0xf72f1000505833571b00000000000000_
        >(the data is always the same for each warcraft pc)
 
     2. **"Server waiting"**  
     When you open a LAN game, the server sends every 5 seconds (may depend on the patch version) a notifcation boradcast:
        * Source: _local IP of server_
        * Destination: _255.255.255.255_
        * Port: _UDP 6112_ 
        * Data: _0xf7321000010000000100000003000000_  
        The data is defined as: 
        
          \# (byte) | Data | dynamic  |  Description
          --------- | ---- | -------- | -------------
          01        |  f7  | no       | W3 identification (fix)
          02        |  32  | no       | W3 identification (fix)
          03        |  10  | no       | W3 identification (fix)
          04        |  00  | no       | Reserved
          05        |  01  | yes      | Number of opened LAN games since Warcraft started. (here 1)
          06        |  00  | no       | Reserved
          07        |  00  | no       | Reserved
          08        |  00  | no       | Reserved
          09        |  01  | no       | Total number of (joined) players in the game. (here only the server himself)
          10        |  00  | no       | Reserved
          11        |  00  | no       | Reserved
          12        |  00  | no       | Reserved
          13        |  03  | no       | Number of possible players on the map. (here 3)
          14        |  00  | no       | Reserved
          15        |  00  | no       | Reserved
          16        |  00  | no       | Reserved
        
     3. **"Abort"**  
     When you abort the open game:
        * Source: _local IP of client_
        * Destination: _255.255.255.255_
        * Port: _UDP 6112_
        * Data: _0xf733080001000000_

3. Proof of Concept  
Try to inform the game server by sending an unicast instead of broadcast by an external tool. For this PoC we used the software [nping](https://nmap.org/nping/)  
   * Start on the remote computer (server) Warcraft 3.
   * Call nping (C:\Program Files (x86)\Nmap\nping) from command line on the client:
     * ***nping -c 1 --udp --source-port 6112 --dest-port 6112 --source-ip 192.168.1.2 --dest-ip 192.168.1.10 --data f72f1000505833571b00000000000000***  
   
   We got the answer from the server with the information about the open LAN game. So we could join the game.  
  ***The PoC works!*** :thumbsup: :smile:  
