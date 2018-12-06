function $(id) { return document.getElementById(id); }

window.onload = function() {
	initMap('a2');
	initMap('michigan');
	initRouteMap('p2p-search');
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

function makeCordEntry(defaultValue, id) {
	var input = document.createElement('input');
	input.type = "number";
	input.value = defaultValue;
	input.id = id;
	return input;
}

function makeLocationEntry(label, defaultLat, defaultLong, id_prefix, onchange) {
	var container = document.createElement('div');
	container.appendChild(document.createTextNode(label));
	var latE = makeCordEntry(defaultLat, id_prefix + "_lat");
	container.append(latE);
	container.appendChild(document.createTextNode("° N"));
	var longE = makeCordEntry(defaultLong, id_prefix + "_long");
	container.append(longE);
	container.appendChild(document.createTextNode("° W"));

	latE.onchange = longE.onchange = onchange;
	return container;
}

function makeDropdown(labels, values) {
	var select = document.createElement('select');
	for(var i = 0; i < labels.length; i++) {
		var option = document.createElement('option');
		option.value = values[i];
		option.innerText = labels[i];
		select.appendChild(option);
	}
	return select;
}

function initRouteMap(id) {
	// Get initial parameters
	var currentScale = 1, xoffset = 0, yoffset = 0;

	function refresh() {
		var sy = $(id + '_src_lat').value;
		var sx = -$(id + '_src_long').value;
		var dy = $(id + '_dest_lat').value;
		var dx = -$(id + '_dest_long').value;

		img.src = 'shortest-path?size=600&frames=' + frameInput.value + '&src=' + sy + ',' + sx + '&dest=' + dy + ',' + dx + '&algorithm=' + algorithmInput.value + '&zoom=' + currentScale + '&xoffset=' + xoffset + '&yoffset=' + yoffset;

	}

	function zoom(ratio) {
		currentScale *= ratio
		refresh();
	}

	function search() {
		currentScale = 1;
		xoffset = 0;
		yoffset = 0;
		refresh();
	}

	// Units are multiples of currentScale
	function pan(x, y) {
		xoffset += x/currentScale;
		yoffset += y/currentScale;
		refresh();
	}

	// Setup image
	var div = $(id);
	var img = document.createElement('img');
	div.appendChild(img);

	// Map zoom/pan controls
	var controls = document.createElement('form');
	controls.appendChild(makeButton('Left', function() { pan(-0.25, 0); }));
	controls.appendChild(makeButton('Right', function() { pan(0.25, 0); }));
	controls.appendChild(makeButton('Up', function() { pan(0, 0.25); }));
	controls.appendChild(makeButton('Down', function() { pan(0, -0.25); }));
	controls.appendChild(makeButton('+', function() { zoom(1/0.75); }))
	controls.appendChild(makeButton('-', function() { zoom(0.75); }))
	div.appendChild(controls)

	// Search box controls
	controls = document.createElement('div');
	controls.appendChild(makeLocationEntry('Source: ', '42.2808', '83.74', id + '_src', search));
	controls.appendChild(makeLocationEntry('Destination: ', '41.65', '83.53', id + '_dest', search));
	div.appendChild(controls)

	// Frame and animation controls
	controls = document.createElement('div');

	// Number of frames input
	controls.append(document.createTextNode('Frames: '));
	var frameInput = document.createElement('input');
	frameInput.type = 'number';
	frameInput.min = 1;
	frameInput.max = 120;
	frameInput.value = 15;
	frameInput.onchange = refresh;
	controls.appendChild(frameInput);

	// Algorithm controls
	controls.append(document.createTextNode('Algorithm: '));
	var algorithmInput = makeDropdown(['Dijkstra', 'ALT (A*, landmarks, triangle inequality)'], ['dijkstra', 'alt']);
	algorithmInput.onchange = refresh;
	controls.append(algorithmInput)
	div.appendChild(controls);
	search();
}
