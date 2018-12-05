function $(id) { return document.getElementById(id); }

function displaySearch() {
	var sy = $('src_lat').value;
	var sx = -$('src_long').value;
	var dy = $('dest_lat').value;
	var dx = -$('dest_long').value;
	var frames = $('frames').value;
	var algorithm = $('algorithm').value;

	$('route_map').src = 'shortest-path?size=600&frames=' + frames + '&src=' + sy + ',' + sx + '&dest=' + dy + ',' + dx + '&algorithm=' + algorithm;
}

window.onload = function() {
	initMap('a2');
	initMap('michigan');

	$('src_lat').onchange = $('src_long').onchange = $('dest_lat').onchange = $('dest_long').onchange = $('frames').onchange = $('algorithm').onchange = displaySearch;
	displaySearch();
}

function makeButton(value, onclick) {
	var button = document.createElement('input');
	button.type = 'button';
	button.value = value;
	button.onclick = onclick;
	return button;
}

function initMap(id) {
	var div = $(id);
	var img;
	for(var i = 0; i < div.childNodes.length; i++) {
		if(div.childNodes[i].className === "map") {
			img = div.childNodes[i];
		}
	}
	// Get initial parameters
	var params = new URLSearchParams(img.src.split('?')[1]);
	var x = parseFloat(params.get('centerx'));
	var y = parseFloat(params.get('centery'));
	var radius = parseFloat(params.get('radius'));

	function refresh() {
		img.src = 'map?centerx=' + x + '&centery=' + y + '&radius=' + radius + '&size=500';
		// TODO: adjust size based on screen
		// TOOD: resize images when browser window changes
	}

	function zoom(ratio) {
		radius *= ratio;
		refresh();
	}

	// Units are as a factor of radius
	function pan(dx, dy) {
		x += dx * radius;
		y += dy * radius;
		refresh();
	}

	var controls = document.createElement('form');
	controls.appendChild(makeButton('Left', function() { pan(-0.25, 0); }));
	controls.appendChild(makeButton('Right', function() { pan(0.25, 0); }));
	controls.appendChild(makeButton('Up', function() { pan(0, 0.25); }));
	controls.appendChild(makeButton('Down', function() { pan(0, -0.25); }));
	controls.appendChild(makeButton('+', function() { zoom(0.75); }))
	controls.appendChild(makeButton('-', function() { zoom(1/0.75); }))
	div.appendChild(controls)
}
