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
  -credentials="admin:admin": credentials specify as <login>:<pass>
  -csvfile="output.csv": file to write csv output to (only used with -daemon)
  -daemon=false: run as a (foreground) daemon with polling webserver
  -debug=false: show debug information
  -do="info": enable/disable/info/disableAP/uptime/reboot
  -info="W": W/E/V/I
                W = centiWatt
                E = milliWatts/h
                V = milliVolts
                I = milliAmps
  -ip="192.168.8.74": ipv4 address of smartplug device
  -port=8080: webserver port (only used with -daemon)
  -raw="": raw command to execute on device (via http)
  -rawt="": raw command to execute on device (via telnet)
```

-do enable / disable: to enable/disable the power output of the plug  
-do disableAP: to disable AP mode on the smartplug (for security reasons)  
-do uptime: show uptime of the device  
-do reboot: reboots the device (does not impacts the power output)  
-do info: get information about the power status (needs -info option)  
   
-raw "command": executes a command on the plug (it's running busybox/linux)(via http)
   > sending commands via http is limited, only one command is possible, can't chain commands)
-rawt "command": executes a command on the plug (it's running busybox/linux)  
   > here you can chain commands. eg command1 && command2 

-daemon: starts a webserver and polls the device every second for information  
   > - port: specify listen port for the webserver (default 8080) (only needed with -daemon)
   > - csvfile: specify cvsfile to write to (default "output.csv") (only needed with -daemon)

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
$ plugctl -ip 192.168.1.50 -credentials "admin:test" -do enable
enabling plug.
```

Get centiWatt usage information about plug on ip 192.168.1.50 with default password
```
$ plugctl -ip 192.168.1.50 -info W -do info
1058 W
```

Disable the AP mode on the smartplug (for security reasons). This is not saved on reboot!
```
$ plugctl -do disableAP
disabling AP.
```

View the CPU info of the device by using the raw command
```
$ plugctl -raw="cat /proc/cpuinfo"
executing command: cat   /proc/cpuinfo

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
$ plugctl -rawt "GetInfo W && GetInfo I && GetInfo E && GetInfo V"
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
