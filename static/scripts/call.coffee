if not window.goule?
	window.goule = {}

window.goule.api = (name, object, callback) ->
	path = window.location.pathname
	match = /^(.*)index.html$/.exec path
	path = match[1] if match?
	$.ajax "#{path}api/#{name}",
		type: 'POST'
		data: JSON.stringify object
		contentType: 'application/json'
		cache: false
		dataType: 'json'
		error: -> callback 'Error making API call.', null
		success: (data) -> callback null, data
	return
