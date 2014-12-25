window.goule = {} if not window.goule?

getSessionId = ->
  # TODO: do this in a better way
  match = /goule_id=(.*)/.exec document.cookie
  return match[1] if match?
  return ''

setSessionId = (id) ->
  # TODO: do this in a better way
  document.cookie = 'goule_id=' + id

singleValue = (name, args, cb) ->
  window.goule.api name, args, (err, x) ->
    return cb err, null if err?
    cb null, x[0]

window.goule.api = (name, args, cb) ->
  argsStrs = (JSON.stringify x for x in args)
  authStr = if name is 'Auth' then '' else '?id=' + getSessionId()
  $.ajax "/api/#{name}#{authStr}",
    type: 'POST'
    data: JSON.stringify argsStrs
    contentType: 'application/json'
    cache: false
    dataType: 'json'
    error: -> cb 'Error making API call.', null
    success: (data) -> cb null, data
  return

window.goule.auth = (password, cb) ->
  singleValue 'Auth', [password], (err, id) ->
    return cb err, null if err?
    return cb null, false if id == ''
    setSessionId id
    cb null, true

window.goule.config = (cb) -> singleValue 'Config', [], cb
