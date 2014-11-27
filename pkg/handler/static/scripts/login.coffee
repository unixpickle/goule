window.goule = {} if not window.goule?

window.goule.login = {}

window.goule.login.show = (animate) ->
  if animate
    $('#login').css display: 'inline-block', opacity: '0'
    $('#login').fadeIn()
  else
    $('#login').css display: 'inline-block', opacity: '1.0'

window.goule.login.hide = (animate) ->
  if animate
    $('#login').fadeOut()
  else
    $('#login').css display: 'none'

$ ->
  $('#login-input').focus()
  $('#login-form').submit (e) ->
    e.preventDefault()
    $('#login-input').prop 'disabled', true
    window.goule.auth $('#login-input').val(), (succ) ->
      $('#login-input').prop 'disabled', false
      if not succ
        $('#login-input').effect 'shake'
      else
        window.goule.login.hide true
        window.goule.headerControls.show true
        window.goule.headerControls.selectTab 0, false