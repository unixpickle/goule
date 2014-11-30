window.goule = {} if not window.goule?

window.goule.api = {}
window.goule.api.run = (name, object, callback) ->
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

window.goule.api.runBool = (name, object, callback) ->
  window.goule.api.run name, object, (err, obj) -> callback not err?

window.goule.api.auth = (password, callback) ->
  window.goule.api.runBool 'auth', password, callback

window.goule.api.listServices = (callback) ->
  window.goule.api.run 'services', null, callback

window.goule.api.changePassword = (newPassword, callback) ->
  window.goule.api.runBool 'change_password', newPassword, callback

window.goule.api.setHttp = (settings, callback) ->
  window.goule.api.runBool 'set_http', settings, callback

window.goule.api.setHttps = (settings, callback) ->
  window.goule.api.runBool 'set_https', settings, callback

window.goule.api.getConfig = (callback) ->
  window.goule.api.run 'get_configuration', null, callback

window.goule.api.setProxy = (settings, callback) ->
  window.goule.api.runBool 'set_proxy', settings, callback

window.goule.api.setSessionTimeout = (timeout, callback) ->
  window.goule.api.runBool 'set_admin_session_timeout', timeout, callback
