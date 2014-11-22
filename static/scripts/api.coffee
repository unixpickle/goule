window.goule = {} if not window.goule?

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

window.goule.boolApi = (name, object, callback) ->
	window.goule.api name, object, (err, obj) -> callback not err?

window.goule.auth = (password, callback) ->
	window.goule.boolApi 'auth', password, callback

window.goule.listServices = (callback) ->
	window.goule.api 'services', null, callback

window.goule.changePassword = (newPassword, callback) ->
	window.goule.boolApi 'change_password', newPassword, callback

window.goule.setHttp = (enabled, port, callback) ->
	window.goule.boolApi 'set_http', {enabled: enabled, port: port}, callback

window.goule.setHttps = (enabled, port, callback) ->
	window.goule.boolApi 'set_https', {enabled: enabled, port: port}, callback
