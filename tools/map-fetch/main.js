var request = require('request');
var fs = require('fs');

var uri_root = 'http://punter.inf.ed.ac.uk/graph-viewer/';
var fs_root = '../';

request(uri_root + 'maps.json', function(error, response, body) {
	var json = JSON.parse(body);
	for (var map of json.maps.concat(json.other_maps)) {
		request(uri_root + map.filename)
		.pipe(fs_root + fs.createWriteStream(map.filename));
	}
});