package main

const (
	plugEnable    = "GpioForCrond+1"
	plugDisable   = "GpioForCrond+0"
	plugInfo      = "GetInfo+"
	plugDisableAP = "ifconfig+ra0+down"
	plugIfconfig  = "ifconfig"
	plugUptime    = "uptime"
	plugReboot    = "reboot"

	plugGetInfoStats = "GetInfo I && GetInfo W && GetInfo E && GetInfo V"

	plugURI        = "/goform/SystemCommand?command="
	plugReadResult = "/adm/system_command.asp"
)

const webHistory = `<html>
	<script src="http://cdnjs.cloudflare.com/ajax/libs/dygraph/1.1.0/dygraph-combined.js"></script>
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.3/jquery.min.js"></script>
	<body>
	<div id="graphdiv" style="width:600px; height:320px;"></div>
	<script type="text/javascript">
	  g = new Dygraph(
		document.getElementById("graphdiv"),
		"/read.csv",{
		showRoller: false,
		//drawPoints: true,
		labels: ["time","Ampere","Watt","Watt/h","Volt"]
		});
	</script>
	</body>
	</html>`

const webStream = `<html>
	<script src="http://cdnjs.cloudflare.com/ajax/libs/dygraph/1.1.0/dygraph-combined.js"></script>
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.3/jquery.min.js"></script>
	<script src="http://cdnjs.cloudflare.com/ajax/libs/jquery-csv/0.71/jquery.csv-0.71.min.js"></script>
	<body>
	<div id="graphdiv" style="width:600px; height:320px;"></div>
	<script type="text/javascript">
	var d = [];
	/*
	$.get( "/read.csv", function( data ) {
		d = $.csv.toArrays(data);
		for (i = 0; i < d.length; i++) {
	    	d[i][0] = new Date(d[i][0]);
		}
	});
	*/

	g = new Dygraph(
	    document.getElementById("graphdiv"),
	    d, {
	    // rollPeriod: 2,
	    showRoller: false,
	    // drawPoints: true,
	    labels: ["time","Ampere","Watt","Watt/h","Volt"]
	    });

	window.intervalId = setInterval(function () {
		 	$.getJSON('/read.json', function(data) {
		 	d.push([new Date(),data[1],data[2],data[3],data[4]])
		    g.updateOptions( { 'file': d } );
	        });
	      }, 1000);
	</script>
	</body>
	</html>`
