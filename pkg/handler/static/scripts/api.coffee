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
  window.goule.api.runBool 'Auth', password, callback

window.goule.api.listServices = (callback) ->
  window.goule.api.run 'ListServices', null, callback

window.goule.api.changePassword = (newPassword, callback) ->
  window.goule.api.runBool 'ChangePassword', newPassword, callback

window.goule.api.setHttp = (settings, callback) ->
  window.goule.api.runBool 'SetHTTP', settings, callback

window.goule.api.setHttps = (settings, callback) ->
  window.goule.api.runBool 'SetHTTPS', settings, callback

window.goule.api.getConfig = (callback) ->
  window.goule.api.run 'GetConfiguration', null, callback

window.goule.api.setProxy = (settings, callback) ->
  window.goule.api.runBool 'SetProxy', settings, callback

window.goule.api.setSessionTimeout = (timeout, callback) ->
  window.goule.api.runBool 'SetAdminSessionTimeout', timeout, callback

window.goule.api.setAdminRules = (rules, callback) ->
  window.goule.api.runBool 'SetAdminRules', rules, callback
