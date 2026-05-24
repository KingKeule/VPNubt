# VPNubt (VPN-udp-broadcast-tunneler)
Our tool "copies" udp broadcasts on the selected port to udp unicasts which are sent to the specific IP address to bypass the VPN router barrier.

## Background
We love to play old school games like Warcraft 3 with friends. 
Since we can't do a LAN session like in our youth, we play over the internet via VPN without using Battle.Net.
The problem with e.g. Warctaft 3 is that the server could not be found even if we are connected via VPN.
> (VPN means here classic OSI layer 3 VPNs and not a OSI layer 2 bridge VPN.) 

## What is the reason for that?
The game server sends an udp broadcast to notify all player in the LAN. When you play over internet via VPN there is normaly a consumer router which does not relay this broadcast otherwise the network/internet would be flooded.
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
   * You can divide it in 5 Parts:

     1. **"Hello" information"**  
     When you enter the network lobby, Warcraft will only send a notifcation broadcast ***once***:
        * Source: _local IP of client_
        * Destination: _255.255.255.255_
        * Port: _UDP 6112_
        * Data: _0xf72f1000505833571b00000000000000_
        >(the data is always the same for each warcraft pc)
  
     2. **"Server created"**
     When you open a LAN game, the server sends a request similar to the "Hello" message **once**:
        * Source: _local IP of the host_
        * Destination: _255.255.255.255_
        * Port: _UDP 6112_
        * Data: _f7311000505833571b00000001000000_
        >only the last 4 bytes change, which is the number of opened LAN games since Warcraft started for that client,
        >most likely working as a local id.
 
     4. **"Server waiting"**  
     After you open a LAN game, the server sends every 5 seconds (may depend on the patch version) a notifcation boradcast:
        * Source: _local IP of server_
        * Destination: _255.255.255.255_
        * Port: _UDP 6112_ 
        * Data: _0xf7321000010000000100000003000000_  
        The data is defined as: 
        
          \# (byte) | Data   | dynamic  |  Description
          --------- | ------ | -------- | -------------
          01        |  f7    | no       | W3 identification (fixed)
          02        |  32    | no       | W3 op-code (0x32 = server waiting)
          03-04     |  10    | no       | Total length of the payload, unsigned short in little endian order (here 16 bytes including this header)
          05-08     |  01    | yes      | Number of opened LAN games since Warcraft started, unsigned int in little endian order. (here 1)
          09-12     |  01    | no       | Total number of (joined) players in the game, unsigned int in little endian order. (here only the server himself)
          13-16     |  03    | no       | Number of possible players on the map, unsigned int in little endian order. (here 3)
        
     5. **"Abort"**  
     When you abort the open game:
        * Source: _local IP of client_
        * Destination: _255.255.255.255_
        * Port: _UDP 6112_
        * Data: _0xf733080001000000_
        >the number of created LAN games/id is also sent at the last 4 bytes

   <details>
   All the messages have the format:
     
   * Starting byte _0xf7_ as a magic number to identify W3 messages
   * Op-code going from _0x2f_ to _0x33_ indicating what is the content of the message
   * 16-bit unsigned int in little endian order to indicate the total length of the message (this header included)
   * The payload which has a byte-length given by the length value minus the header length (4 bytes)
  
   ---
    
   There is also a sixth type of UDP message starting with _0xf730..._ which sends the entire lobby information which includes the name of the host
   and it's sent to any peers sending the "Hello" message. Not included in this readme, but you can use this information if you want to display it in
   an external program.
   </details>

3. Proof of Concept  
Try to inform the game server by sending an unicast instead of broadcast by an external tool. For this PoC we used the software [nping](https://nmap.org/nping/)  
   * Start on the remote computer (server) Warcraft 3.
   * Call nping (C:\Program Files (x86)\Nmap\nping) from command line on the client:
     * ***nping -c 1 --udp --source-port 6112 --dest-port 6112 --source-ip 192.168.1.2 --dest-ip 192.168.1.10 --data f72f1000505833571b00000000000000***  
   
   We got the answer from the server with the information about the open LAN game. So we could join the game.  
  ***The PoC works!*** :thumbsup: :smile:  
