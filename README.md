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
```

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


