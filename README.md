# plugctl

Control your Smartplug from Maginon / Aldi  
Based upon information from https://www.dealabs.com/bons-plans/prise-wifi-/85521?page=36 and https://github.com/netdata/loxone/tree/master/maginon_Smart-Plug

## Usage
```
$ plugctl
  -credentials="admin:admin": credentials specify as <login>:<pass>
  -do="": enable/disable/info/disableAP/uptime/reboot
  -info="": W/E/V/I
                W = centiWatt
                E = milliWatts/h
                V = milliVolts
                I = milliAmps
  -ip="192.168.8.74": ipv4 address of smartplug device
  -raw="": raw command to execute
```

-do enable / disable: to enable/disable the power output of the plug  
-do disableAP: to disable AP mode on the smartplug (for security reasons)  
-do uptime: show uptime of the device  
-do reboot: reboots the device (does not impacts the power output)  
-do info: get information about the power status (needs -info option)  
   
-raw "command": executes a command on the plug (it's running busybox/linux)  

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
$ plugctl.exe -raw="cat /proc/cpuinfo"
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

