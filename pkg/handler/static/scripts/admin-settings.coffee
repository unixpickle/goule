window.goule = {} if not window.goule?

class AdminSettings
  show: (animate) ->
    container = $ '#admin-settings-container'
    if animate
      container.css display: 'inline-block', opacity: '0.0'
      container.animate opacity: 1.0
    else
      container.css display: 'inline-block', opacity: '1.0'
  
  hide: (animate) ->
    container = $ '#admin-settings-container'
    if animate
      container = $('#admin-settings-container')
      $('#admin-settings-container').fadeOut()
    else
      container.css display: 'none'

window.goule.adminSettings = new AdminSettings()

chpassSubmit = ->
  newPass = $('#new-password').val()
  confirm = $('#confirm-password').val()
  if newPass isnt confirm
    $('#confirm-password').effect 'shake'
    return
  $('#change-password-button').prop 'disabled', true
  window.goule.changePassword newPass, ->
    $('#change-password-button').prop 'disabled', false
    $('#new-password').val ''
    $('#confirm-password').val ''

$ ->
  $('#change-password-button').click chpassSubmit
  $('#change-password-form').click (e) ->
    e.preventDefault()
    chpassSubmit()
    
