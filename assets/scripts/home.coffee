$ ->
  window.goule.config (err, data) ->
    window.location = 'login' if err?
    handleConfig data

handleConfig = (data) ->
  