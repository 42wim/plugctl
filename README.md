# plugctl

Control your Smartplug from Maginon / Aldi  
Based upon information from https://www.dealabs.com/bons-plans/prise-wifi-/85521?page=36 and https://github.com/netdata/loxone/tree/master/maginon_Smart-Plug

## Build
Install Go using your package manager or from the website https://golang.org/doc/install  
Download the plugctl source or use git 

Example

```
$ git clone https://github.com/42wim/plugctl.git
$ cd plugctl
$ go build
```

You will now have the plugctl executable in the same directory


## Usage
```
$ plugctl
  -conf="": a valid config file (uses plugctl.conf if exists)
  -credentials="admin:admin": credentials specify as <login>:<pass>
  -csvfile="output.csv": file to write csv output to (only used with -daemon)
  -daemon=false: run as a (foreground) daemon with polling webserver
  -debug=false: show debug information
  -delay=1: polling delay of statistics in seconds (only used with -daemon)
  -disable="": disable power/cloud/ap
  -enable="": enable power/cloud/ap
  -ip="192.168.8.74": ipv4 address of smartplug device
  -port=8080: webserver port (only used with -daemon)
  -raw="": raw command to execute on device (via telnet)
  -show="": show info/uptime
```

- enable
  * power : enable power output of the plug  
  * ap: disable AP mode on the smartplug (for security reasons). Also saved in NVRAM (survives reboot/powerfailure)  
  * cloud: disable connections from smartplug to the cloud (iotc). Does not survive a reboot/powerfailure!  

- disable: opposite of enable options above  

- show
  * info: shows current Ampere - Watt - Watt/hour - Volt usage  
  * uptime: show uptime of the device
   
- raw "command": executes a command on the plug (it's running busybox/linux)  
   > here you can chain commands. eg command1 && command2 

- daemon: starts a webserver and polls the device every second for information  
   > - port: specify listen port for the webserver (default 8080) (only needed with -daemon)
   > - csvfile: specify cvsfile to write to (default "output.csv") (only needed with -daemon)

## Configfile
See plugctl.conf.sample (https://github.com/42wim/plugctl/blob/master/plugctl.conf.sample)

If plugctl.conf exists in the current directory it will be used, otherwise a config file can specified using the -conf flag  

E.g. The ip of your plug can be specified in plugctl.conf, so you don't need to give the -ip option with every command  


## Webserver
When -daemon option is used, a webserver will listen by default on port 8080

### Available URL
  * /history - webpage/javascript which parses the csvfile history + current data into a chart
    ![history](http://snag.gy/629gM.jpg)

  * /stream - webpage/javascript which shows a realtime chart
    ![stream](http://snag.gy/dCYY0.jpg)

## Examples
Enable plug on ip 192.168.1.50 with login admin and password test

```
$ plugctl -ip 192.168.1.50 -credentials "admin:test" -enable power
enabling plug.
```

Get usage information about plug on ip 192.168.1.50 with default password
```
$ plugctl -ip 192.168.1.50 -show info
0.01 Ampere - 0.07 Watt - 0.00 Watt/hour - 230.90 Volt
```

Disable the AP mode on the smartplug (for security reasons). This is saved on reboot!
```
$ plugctl -disable ap
disabling AP...success
saving state...already set
```

View the CPU info of the device by using the raw command
```
$ plugctl -raw="cat /proc/cpuinfo"
system type             : Ralink SoC
processor               : 0
cpu model               : MIPS 24K V4.12
BogoMIPS                : 239.61
wait instruction        : yes
microsecond timers      : yes
tlb_entries             : 32
extra interrupt vector  : yes
hardware watchpoint     : yes
ASEs implemented        : mips16 dsp
VCED exceptions         : not available
VCEI exceptions         : not available
```

Get all the usage information Watt/Ampere/Energy/Volt in one go using the rawt command
```
$ plugctl -raw "GetInfo W && GetInfo I && GetInfo E && GetInfo V"
$01W00 000007
$01I00 000064
$01E00 002134
$01V00 236728
```

Start daemon/webserver on port 8888 with debug and device on ip 192.168.1.50 and save CSVfile to plug.csv

```
$ plugctl -daemon -debug -port 8888 -ip 192.168.1.50 -csvfile plug.csv
starting foreground daemon ;-)
[2015/01/11 21:28:13 3.84 66.37 4.47 233.06] took 392.1955ms
[2015/01/11 21:28:14 3.82 66.37 4.47 233.12] took 366.52ms
```

You can now surf to http://localhost:8888/stream to see realtime chart updating.
If you're running for a while or have historic data in plug.csv you can go to http://localhost:8888/history
